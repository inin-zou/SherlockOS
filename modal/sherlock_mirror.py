"""
SherlockOS - HunyuanWorld-Mirror 3D Reconstruction Service
Deployed on Modal with A100 40GB GPU.

Receives crime scene images via HTTP POST, runs HunyuanWorld-Mirror inference
to produce 3D Gaussian Splatting output (gaussians.ply), uploads it to Supabase
Storage, and returns the result as JSON compatible with the Go backend.

Endpoint: https://ykzou1214--sherlock-mirror-reconstruct.modal.run
"""

import io
import os
import struct
import time
import uuid
import base64
import traceback
from typing import Optional

import modal

# ---------------------------------------------------------------------------
# Modal image definition
# ---------------------------------------------------------------------------

mirror_image = (
    modal.Image.debian_slim(python_version="3.10")
    .apt_install("git", "wget", "libgl1-mesa-glx", "libglib2.0-0", "ffmpeg")
    .pip_install(
        "torch==2.4.1",
        "torchvision==0.19.1",
        "numpy<2",
        "Pillow>=10.0.0",
        "plyfile>=1.0.0",
        "huggingface-hub>=0.20.0",
        "safetensors>=0.4.0",
        "einops>=0.7.0",
        "timm>=0.9.0",
        "requests>=2.31.0",
        "tqdm>=4.65.0",
        "scipy>=1.11.0",
        "trimesh>=4.0.0",
        "httpx>=0.25.0",
        "roma>=1.5.0",
        "opencv-python-headless>=4.8.0",
    )
    # Install gsplat from PyPI (pre-built CUDA wheels, avoids needing CUDA toolkit for compilation)
    .pip_install("gsplat>=1.0.0")
    # Install HunyuanWorld-Mirror from source (not pip-installable, clone + add to PYTHONPATH)
    .run_commands(
        "git clone --recursive https://github.com/Tencent-Hunyuan/HunyuanWorld-Mirror.git /root/HunyuanWorld-Mirror",
        "pip install -r /root/HunyuanWorld-Mirror/requirements.txt || true",
    )
    .env({"PYTHONPATH": "/root/HunyuanWorld-Mirror"})
    # Pre-download model weights during image build
    .run_commands(
        "python -c \""
        "from huggingface_hub import snapshot_download; "
        "snapshot_download('tencent/HunyuanWorld-Mirror', local_dir='/root/model_weights')"
        "\""
    )
)

ply_volume = modal.Volume.from_name("sherlock-ply-storage", create_if_missing=True)

app = modal.App("sherlock-mirror", image=mirror_image)

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

MODEL_DIR = "/root/model_weights"
PLY_VOLUME_PATH = "/ply-storage"
TARGET_SIZE = 518  # HunyuanWorld-Mirror input resolution
MAX_POINTS_IN_RESPONSE = 50_000  # Cap point cloud size in JSON response
SUPABASE_BUCKET = "case-assets"


# ---------------------------------------------------------------------------
# Helper: Supabase upload
# ---------------------------------------------------------------------------

def upload_to_supabase(data: bytes, asset_key: str, content_type: str = "application/octet-stream") -> None:
    """Upload binary data to Supabase Storage via REST API."""
    import requests as req

    supabase_url = os.environ.get("SUPABASE_URL", "")
    supabase_key = os.environ.get("SUPABASE_SECRET_KEY", "")

    if not supabase_url or not supabase_key:
        raise RuntimeError("SUPABASE_URL and SUPABASE_SECRET_KEY must be set")

    url = f"{supabase_url}/storage/v1/object/{SUPABASE_BUCKET}/{asset_key}"
    headers = {
        "Authorization": f"Bearer {supabase_key}",
        "Content-Type": content_type,
        "x-upsert": "true",
    }

    resp = req.post(url, headers=headers, data=data, timeout=120)
    if resp.status_code not in (200, 201):
        raise RuntimeError(
            f"Supabase upload failed ({resp.status_code}): {resp.text}"
        )


# ---------------------------------------------------------------------------
# Helper: PLY writer for Gaussian Splatting output
# ---------------------------------------------------------------------------

def write_gaussian_ply(
    path: str,
    means: "numpy.ndarray",
    opacities: "numpy.ndarray",
    scales: "numpy.ndarray",
    quats: "numpy.ndarray",
    sh: "numpy.ndarray",
    colors_rgb: Optional["numpy.ndarray"] = None,
) -> None:
    """
    Write a .ply file encoding 3D Gaussian Splatting parameters.

    Parameters
    ----------
    means : (N, 3) float32 - Gaussian centres
    opacities : (N, 1) or (N,) float32 - opacity (logit or raw)
    scales : (N, 3) float32 - log-scale of each axis
    quats : (N, 4) float32 - rotation quaternion (w, x, y, z)
    sh : (N, C) float32 - spherical harmonic coefficients (at least dc term)
    colors_rgb : optional (N, 3) float32 in [0, 1]
    """
    import numpy as np

    n = means.shape[0]
    opacities = opacities.reshape(n, -1)
    if opacities.shape[1] != 1:
        opacities = opacities[:, :1]

    # Determine SH degree from coefficient count
    sh_dim = sh.shape[1] if sh.ndim == 2 else 1
    # Ensure 3-channel SH: total coefficients should be divisible by 3
    if sh_dim % 3 == 0:
        n_sh_per_channel = sh_dim // 3
    else:
        # Treat as single-channel (dc only) and replicate
        sh = np.tile(sh.reshape(n, -1), (1, 3))
        n_sh_per_channel = sh.shape[1] // 3

    # Build PLY header
    header_lines = [
        "ply",
        "format binary_little_endian 1.0",
        f"element vertex {n}",
        "property float x",
        "property float y",
        "property float z",
        "property float nx",
        "property float ny",
        "property float nz",
    ]
    # SH coefficients: f_dc_0..2 then f_rest_0..N
    for i in range(3):
        header_lines.append(f"property float f_dc_{i}")
    rest_count = n_sh_per_channel - 1
    for i in range(rest_count * 3):
        header_lines.append(f"property float f_rest_{i}")

    header_lines.append("property float opacity")
    for i in range(3):
        header_lines.append(f"property float scale_{i}")
    for i in range(4):
        header_lines.append(f"property float rot_{i}")
    header_lines.append("end_header")
    header = "\n".join(header_lines) + "\n"

    # Prepare SH data: split into dc and rest
    sh_flat = sh.reshape(n, 3, n_sh_per_channel)  # (N, 3, K)
    f_dc = sh_flat[:, :, 0]  # (N, 3)
    if n_sh_per_channel > 1:
        f_rest = sh_flat[:, :, 1:].reshape(n, -1)  # (N, 3*(K-1))
    else:
        f_rest = np.zeros((n, 0), dtype=np.float32)

    normals = np.zeros((n, 3), dtype=np.float32)

    with open(path, "wb") as f:
        f.write(header.encode("ascii"))
        for i in range(n):
            # xyz
            f.write(struct.pack("<3f", *means[i]))
            # normals
            f.write(struct.pack("<3f", *normals[i]))
            # f_dc
            f.write(struct.pack(f"<3f", *f_dc[i]))
            # f_rest
            if f_rest.shape[1] > 0:
                f.write(struct.pack(f"<{f_rest.shape[1]}f", *f_rest[i]))
            # opacity
            f.write(struct.pack("<f", float(opacities[i, 0])))
            # scale
            f.write(struct.pack("<3f", *scales[i]))
            # rotation
            f.write(struct.pack("<4f", *quats[i]))


# ---------------------------------------------------------------------------
# Helper: Extract point cloud colours from SH dc term
# ---------------------------------------------------------------------------

def sh_dc_to_rgb(sh_dc: "numpy.ndarray") -> "numpy.ndarray":
    """
    Convert zeroth-order SH coefficient to RGB in [0, 1].
    SH_C0 = 0.28209479177387814
    color = sh_dc * SH_C0 + 0.5
    """
    import numpy as np

    SH_C0 = 0.28209479177387814
    rgb = sh_dc * SH_C0 + 0.5
    return np.clip(rgb, 0.0, 1.0)


# ---------------------------------------------------------------------------
# Health check endpoint
# ---------------------------------------------------------------------------

@app.function(
    gpu=None,
    timeout=30,
)
@modal.fastapi_endpoint(method="GET")
def health():
    """Health check -- confirms the app is deployed and responsive."""
    return {
        "status": "ok",
        "service": "sherlock-mirror",
        "model": "HunyuanWorld-Mirror",
        "gpu": "A100-40GB",
    }


# ---------------------------------------------------------------------------
# PLY download endpoint
# ---------------------------------------------------------------------------

@app.function(
    gpu=None,
    timeout=60,
    volumes={PLY_VOLUME_PATH: ply_volume},
)
@modal.fastapi_endpoint(method="GET")
def download_ply(filename: str):
    """Download a PLY file from the volume by filename."""
    from fastapi.responses import FileResponse, JSONResponse

    ply_volume.reload()
    filepath = f"{PLY_VOLUME_PATH}/{filename}"
    if not os.path.exists(filepath):
        return JSONResponse(status_code=404, content={"error": "PLY not found"})
    return FileResponse(filepath, media_type="application/octet-stream", filename=filename)


# ---------------------------------------------------------------------------
# Main reconstruction endpoint
# ---------------------------------------------------------------------------

@app.function(
    gpu="A100-80GB",
    timeout=300,
    memory=32768,
    secrets=[modal.Secret.from_name("supabase-secrets")],
    scaledown_window=120,
    volumes={PLY_VOLUME_PATH: ply_volume},
)
@modal.fastapi_endpoint(method="POST")
def reconstruct(item: dict):
    """
    Receive crime-scene images OR a video, run HunyuanWorld-Mirror 3D Gaussian
    Splatting reconstruction, upload the .ply to Supabase, and return a JSON
    response that is backward-compatible with the Go backend.

    Request body
    ------------
    {
        "case_id": "<uuid>",
        "scan_asset_keys": ["<base64_img>", ...],   // base64 images (used when video_base64 is absent)
        "video_base64": "<base64_mp4>",              // optional: base64-encoded video file
        "camera_poses": null,
        "existing_scenegraph": null
    }

    When ``video_base64`` is provided the endpoint extracts frames from the
    video at 1 FPS using HunyuanWorld-Mirror's built-in extraction utility and
    uses them for reconstruction instead of ``scan_asset_keys``.
    """
    import numpy as np
    import torch
    from PIL import Image
    from pathlib import Path

    start_ms = int(time.time() * 1000)

    # ------------------------------------------------------------------
    # 1. Parse and validate request
    # ------------------------------------------------------------------
    case_id = item.get("case_id")
    if not case_id:
        return _error_response("case_id is required", status=400)

    video_b64 = item.get("video_base64")
    b64_images = item.get("scan_asset_keys", [])

    # At least one source is required
    if not video_b64 and (not b64_images or len(b64_images) == 0):
        return _error_response(
            "Either video_base64 or scan_asset_keys with at least one base64 image is required",
            status=400,
        )

    use_video = video_b64 is not None and len(video_b64) > 0
    print(
        f"[reconstruct] case_id={case_id}, "
        f"mode={'video' if use_video else 'images'}, "
        f"images={len(b64_images) if not use_video else 'N/A (video)'}"
    )

    # ------------------------------------------------------------------
    # 2. Decode and preprocess input (images or video)
    # ------------------------------------------------------------------
    pil_images = []  # populated by either video or image path

    if use_video:
        # ---- VIDEO PATH ------------------------------------------------
        # Extract frames ourselves using OpenCV, then feed as PIL images
        # (HunyuanWorld-Mirror's extract_load_and_preprocess_images can be
        # unreliable when video_to_image_frames fails silently)
        try:
            import cv2

            video_bytes = base64.b64decode(video_b64)
            video_path = f"/tmp/{case_id}_input.mp4"
            with open(video_path, "wb") as vf:
                vf.write(video_bytes)
            print(f"[reconstruct] Wrote video to {video_path} ({len(video_bytes)} bytes)")

            # Extract frames at 1 FPS using OpenCV
            cap = cv2.VideoCapture(video_path)
            if not cap.isOpened():
                raise RuntimeError(f"OpenCV could not open video: {video_path}")

            src_fps = cap.get(cv2.CAP_PROP_FPS) or 30.0
            total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
            frame_interval = max(1, int(src_fps / 1))  # 1 FPS extraction
            print(f"[reconstruct] Video: {src_fps:.1f} FPS, {total_frames} frames, extracting every {frame_interval} frames")

            frame_idx = 0
            extracted = 0
            max_frames = 100  # Safety cap
            while True:
                ret, frame = cap.read()
                if not ret:
                    break
                if frame_idx % frame_interval == 0 and extracted < max_frames:
                    # Convert BGR → RGB → PIL
                    rgb = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
                    img = Image.fromarray(rgb).resize((TARGET_SIZE, TARGET_SIZE), Image.LANCZOS)
                    pil_images.append(img)
                    extracted += 1
                frame_idx += 1
            cap.release()

            print(f"[reconstruct] Extracted {len(pil_images)} frames from video")

            if len(pil_images) == 0:
                raise RuntimeError("No frames could be extracted from video")

            # Clean up temp video file
            try:
                os.remove(video_path)
            except OSError:
                pass

            # Video path now uses the same image processing as the image path below

        except Exception as exc:
            traceback.print_exc()
            return _error_response(f"Failed to process video input: {exc}", status=400)
    else:
        # ---- IMAGE PATH ------------------------------------------------
        try:
            for idx, b64 in enumerate(b64_images):
                raw = base64.b64decode(b64)
                img = Image.open(io.BytesIO(raw)).convert("RGB")
                # Resize to the model's target resolution
                img = img.resize((TARGET_SIZE, TARGET_SIZE), Image.LANCZOS)
                pil_images.append(img)
                print(f"  decoded image {idx}: {img.size}")
        except Exception as exc:
            return _error_response(f"Failed to decode images: {exc}", status=400)

    # ------------------------------------------------------------------
    # 3. Run HunyuanWorld-Mirror inference
    # ------------------------------------------------------------------
    try:
        print("[reconstruct] Loading model ...")
        from src.models.models.worldmirror import WorldMirror
        from torchvision import transforms

        device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        model = WorldMirror.from_pretrained(MODEL_DIR)
        model = model.to(device)
        model.eval()

        # Both video and image paths produce pil_images; build tensors uniformly
        # HunyuanWorld-Mirror expects [0,1] range tensors with shape [1, S, 3, H, W]
        transform = transforms.Compose([
            transforms.ToTensor(),  # Converts PIL [0,255] → Tensor [0,1]
        ])

        # Stack images and add batch dimension: [S, 3, H, W] → [1, S, 3, H, W]
        img_tensor = torch.stack([transform(img) for img in pil_images]).unsqueeze(0).to(device)
        # cond_flags: list of 3 binary flags [pose_cond, intrinsics_cond, depth_cond]
        cond_flags = [0, 0, 0]

        input_desc = "video frames" if use_video else "images"
        print(f"[reconstruct] Running inference on {len(pil_images)} {input_desc}, tensor shape: {img_tensor.shape} ...")
        with torch.no_grad():
            predictions = model(views={"img": img_tensor}, cond_flags=cond_flags)

        print("[reconstruct] Inference complete")

    except Exception as exc:
        traceback.print_exc()
        return _error_response(f"Model inference failed: {exc}", status=500)

    # ------------------------------------------------------------------
    # 4. Extract Gaussian Splatting parameters
    # ------------------------------------------------------------------
    try:
        # HunyuanWorld-Mirror outputs a dict with gaussian parameters
        # Exact keys depend on the model version; handle common variants
        pts3d = _extract_tensor(predictions, ["pts3d", "means3D", "means", "xyz"])
        conf = _extract_tensor(predictions, ["conf", "confidence", "opacity", "opacities"])
        scales = _extract_tensor(predictions, ["scales", "scaling", "log_scales"])
        quats = _extract_tensor(predictions, ["quats", "rotations", "rotation", "rots"])
        sh_coeffs = _extract_tensor(predictions, ["sh", "shs", "features_dc", "sh_coeffs", "colors_precomp"])

        if pts3d is None:
            return _error_response("Model did not produce point positions (pts3d)", status=500)

        # Move everything to CPU / numpy
        pts3d_np = pts3d.detach().cpu().numpy().reshape(-1, 3).astype(np.float32)
        n_points = pts3d_np.shape[0]
        print(f"[reconstruct] Extracted {n_points} Gaussians")

        # Confidence / opacity
        if conf is not None:
            opacity_np = conf.detach().cpu().numpy().reshape(n_points, -1)[:, :1].astype(np.float32)
        else:
            opacity_np = np.ones((n_points, 1), dtype=np.float32) * 0.8

        # Scales
        if scales is not None:
            scales_np = scales.detach().cpu().numpy().reshape(n_points, 3).astype(np.float32)
        else:
            scales_np = np.full((n_points, 3), -5.0, dtype=np.float32)  # log-scale default

        # Quaternions
        if quats is not None:
            quats_np = quats.detach().cpu().numpy().reshape(n_points, 4).astype(np.float32)
        else:
            quats_np = np.tile(np.array([1, 0, 0, 0], dtype=np.float32), (n_points, 1))

        # SH coefficients
        if sh_coeffs is not None:
            sh_np = sh_coeffs.detach().cpu().numpy().reshape(n_points, -1).astype(np.float32)
        else:
            # Fall back: build dc-only SH from a neutral gray
            sh_np = np.zeros((n_points, 3), dtype=np.float32)

        # Derive RGB colours from SH dc term for the point cloud response
        if sh_np.shape[1] >= 3:
            sh_dc = sh_np[:, :3]
        else:
            sh_dc = np.tile(sh_np[:, :1], (1, 3))
        colors_rgb = sh_dc_to_rgb(sh_dc)

    except Exception as exc:
        traceback.print_exc()
        return _error_response(f"Failed to extract Gaussian parameters: {exc}", status=500)

    # ------------------------------------------------------------------
    # 5. Write gaussians.ply
    # ------------------------------------------------------------------
    try:
        ply_id = str(uuid.uuid4())
        local_ply_path = f"/tmp/gaussians_{ply_id}.ply"

        # Try the model's built-in export first
        exported = False
        if hasattr(predictions, "save_ply") or (isinstance(predictions, dict) and "save_ply" in dir(model)):
            try:
                model.save_ply(local_ply_path, predictions)
                exported = True
                print("[reconstruct] Used model built-in save_ply")
            except Exception:
                pass

        if not exported:
            write_gaussian_ply(
                path=local_ply_path,
                means=pts3d_np,
                opacities=opacity_np,
                scales=scales_np,
                quats=quats_np,
                sh=sh_np,
                colors_rgb=colors_rgb,
            )
            print(f"[reconstruct] Wrote PLY manually ({os.path.getsize(local_ply_path)} bytes)")

    except Exception as exc:
        traceback.print_exc()
        return _error_response(f"Failed to write PLY: {exc}", status=500)

    # ------------------------------------------------------------------
    # 6. Save .ply to Modal Volume (avoids Supabase size limits)
    # ------------------------------------------------------------------
    ply_filename = f"gaussians_{ply_id}.ply"
    volume_ply_path = f"{PLY_VOLUME_PATH}/{ply_filename}"
    gaussian_asset_key = None
    try:
        import shutil
        os.makedirs(PLY_VOLUME_PATH, exist_ok=True)
        shutil.copy2(local_ply_path, volume_ply_path)
        ply_volume.commit()
        gaussian_asset_key = ply_filename  # just the filename; served via /download endpoint
        print(f"[reconstruct] Saved PLY to volume: {volume_ply_path} ({os.path.getsize(local_ply_path)} bytes)")
    except Exception as exc:
        traceback.print_exc()
        print(f"[reconstruct] WARNING: Volume save failed: {exc}")

    # ------------------------------------------------------------------
    # 7. Clean up temp file
    # ------------------------------------------------------------------
    try:
        os.remove(local_ply_path)
    except OSError:
        pass

    # ------------------------------------------------------------------
    # 8. Build backward-compatible JSON response
    # ------------------------------------------------------------------
    elapsed_ms = int(time.time() * 1000) - start_ms

    # Compute bounding box from point positions
    bbox_min = pts3d_np.min(axis=0).tolist()
    bbox_max = pts3d_np.max(axis=0).tolist()
    centroid = pts3d_np.mean(axis=0).tolist()

    # Downsample point cloud for the JSON response if needed
    if n_points > MAX_POINTS_IN_RESPONSE:
        indices = np.random.default_rng(42).choice(n_points, MAX_POINTS_IN_RESPONSE, replace=False)
        pc_positions = pts3d_np[indices].tolist()
        pc_colors = colors_rgb[indices].tolist()
        pc_count = MAX_POINTS_IN_RESPONSE
    else:
        pc_positions = pts3d_np.tolist()
        pc_colors = colors_rgb.tolist()
        pc_count = n_points

    # Scene object representing the full reconstruction
    obj_id = str(uuid.uuid4())
    response = {
        "objects": [
            {
                "id": str(uuid.uuid4()),
                "action": "create",
                "confidence": 0.8,
                "object": {
                    "id": obj_id,
                    "type": "other",
                    "label": "Reconstructed Scene",
                    "pose": {
                        "position": centroid,
                        "rotation": [1.0, 0.0, 0.0, 0.0],
                        "scale": [1.0, 1.0, 1.0],
                    },
                    "bbox": {
                        "min": bbox_min,
                        "max": bbox_max,
                    },
                    "state": "visible",
                    "confidence": 0.8,
                },
                "source_images": [],
            }
        ],
        "point_cloud": {
            "positions": pc_positions,
            "colors": pc_colors,
            "count": pc_count,
        },
        "gaussian_asset_key": gaussian_asset_key,
        "mesh_asset_key": None,
        "pointcloud_asset_key": None,
        "uncertainty_regions": [],
        "processing_stats": {
            "input_images": len(pil_images),
            "detected_objects": 1,
            "point_count": n_points,
            "processing_time_ms": elapsed_ms,
            "input_type": "video" if use_video else "images",
        },
    }

    input_desc = "video" if use_video else f"{len(b64_images)} images"
    print(
        f"[reconstruct] Done. input={input_desc}, points={n_points}, "
        f"elapsed={elapsed_ms}ms, ply_key={gaussian_asset_key}"
    )
    return response


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

def _extract_tensor(predictions, candidate_keys: list):
    """
    Try to pull a tensor out of the model predictions dict (or object)
    using several possible key names.
    """
    if isinstance(predictions, dict):
        for k in candidate_keys:
            if k in predictions:
                return predictions[k]
    else:
        for k in candidate_keys:
            if hasattr(predictions, k):
                return getattr(predictions, k)
    return None


def _error_response(message: str, status: int = 500) -> dict:
    """Return a minimal error payload. Modal web endpoints return 200 by
    default; we embed the error in the body.  For true HTTP error codes,
    raise modal.web_endpoint exceptions, but to keep backward compat
    with the Go client (which reads the body), we log and return JSON."""
    import fastapi

    print(f"[reconstruct] ERROR ({status}): {message}")
    raise fastapi.HTTPException(status_code=status, detail=message)
