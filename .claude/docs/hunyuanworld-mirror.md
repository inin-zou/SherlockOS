# HunyuanWorld-Mirror API Documentation

> **用途**：SherlockOS 场景重建（Reconstruction Worker）
> **GitHub**：https://github.com/Tencent-Hunyuan/HunyuanWorld-Mirror
> **HuggingFace**：https://huggingface.co/tencent/HunyuanWorld-Mirror

## 概述

HunyuanWorld-Mirror 是腾讯混元的通用 3D 几何预测模型，支持从单张或多张图片重建 3D 场景。

### 核心能力

| 能力 | 输出 |
|------|------|
| 点云重建 | Point Cloud (N, 3) |
| 深度估计 | Depth Map (H, W) |
| 相机位姿估计 | Camera Pose (rotation, translation) |
| 法线预测 | Surface Normals (H, W, 3) |
| 新视角合成 | Novel Views (K, 3, H, W) |
| 3D Gaussian | Gaussian Splats |

## 安装

### 环境要求

- Python 3.10+
- CUDA 11.8+ 或 12.x
- GPU: 建议 RTX 4090 或更高（24GB+ VRAM）

### 安装步骤

```bash
# 克隆仓库
git clone https://github.com/Tencent-Hunyuan/HunyuanWorld-Mirror
cd HunyuanWorld-Mirror

# 安装 PyTorch (CUDA 12.1)
pip install torch==2.3.0 torchvision==0.18.0 --index-url https://download.pytorch.org/whl/cu121

# 安装依赖
pip install -r requirements.txt

# 可选：安装 3D Gaussian 优化组件
pip install gsplat
```

### 模型下载

```bash
# 自动下载（运行时）
# 模型会在首次运行时自动从 HuggingFace 下载

# 或手动下载
huggingface-cli download tencent/HunyuanWorld-Mirror --local-dir ./models/HunyuanWorld-Mirror
```

## 基础用法

### 单图推理

```python
import torch
from PIL import Image
from hyworld import HunyuanWorldMirror

# 加载模型
model = HunyuanWorldMirror.from_pretrained("tencent/HunyuanWorld-Mirror")
model = model.cuda().eval()

# 加载图像
image = Image.open("crime_scene_01.jpg")

# 推理
with torch.no_grad():
    outputs = model.predict(image)

# 提取结果
point_cloud = outputs["point_cloud"]      # (N, 3) numpy array
depth_map = outputs["depth"]              # (H, W) numpy array
normals = outputs["normals"]              # (H, W, 3) numpy array
camera_params = outputs["camera"]         # dict: intrinsics, extrinsics
```

### 多图推理（多视角）

```python
import torch
from PIL import Image
from hyworld import HunyuanWorldMirror

model = HunyuanWorldMirror.from_pretrained("tencent/HunyuanWorld-Mirror").cuda().eval()

# 加载多张图像
images = [
    Image.open("scene_view_1.jpg"),
    Image.open("scene_view_2.jpg"),
    Image.open("scene_view_3.jpg"),
]

# 多视角推理
with torch.no_grad():
    outputs = model.predict_multi_view(images)

# 融合后的点云
fused_point_cloud = outputs["fused_point_cloud"]  # (N, 3)
camera_poses = outputs["camera_poses"]            # list of poses
```

### 带先验信息的推理

```python
import torch
import numpy as np
from hyworld import HunyuanWorldMirror

model = HunyuanWorldMirror.from_pretrained("tencent/HunyuanWorld-Mirror").cuda().eval()

# 准备先验信息（如果有）
priors = {
    # 深度先验（来自深度传感器）
    "depth": depth_map,  # (H, W) numpy array

    # 相机内参（如果已知）
    "intrinsics": np.array([
        [fx, 0, cx],
        [0, fy, cy],
        [0, 0, 1]
    ]),

    # 相机位姿（如果已知）
    "pose": {
        "rotation": rotation_matrix,  # (3, 3)
        "translation": translation_vector  # (3,)
    }
}

with torch.no_grad():
    outputs = model.predict(image, priors=priors)
```

## SherlockOS 集成示例

### Reconstruction Worker 实现

```python
import torch
import numpy as np
from PIL import Image
from typing import List, Optional
from hyworld import HunyuanWorldMirror
import trimesh

class ReconstructionWorker:
    def __init__(self, model_path: str = "tencent/HunyuanWorld-Mirror"):
        self.model = HunyuanWorldMirror.from_pretrained(model_path)
        self.model = self.model.cuda().eval()

    def process_scans(
        self,
        image_paths: List[str],
        camera_poses: Optional[List[dict]] = None,
        existing_scenegraph: Optional[dict] = None
    ) -> dict:
        """处理扫描图片，返回场景重建结果"""

        # 加载图片
        images = [Image.open(p) for p in image_paths]

        # 准备先验
        priors = None
        if camera_poses:
            priors = {"camera_poses": camera_poses}

        # 推理
        with torch.no_grad():
            if len(images) == 1:
                outputs = self.model.predict(images[0], priors=priors)
            else:
                outputs = self.model.predict_multi_view(images, priors=priors)

        # 提取对象提案
        objects = self._extract_objects(outputs, image_paths)

        # 保存 mesh（如果需要）
        mesh_path = self._save_mesh(outputs) if "mesh" in outputs else None

        return {
            "objects": objects,
            "mesh_asset_key": mesh_path,
            "pointcloud_asset_key": self._save_pointcloud(outputs),
            "uncertainty_regions": self._compute_uncertainty(outputs),
            "processing_stats": {
                "input_images": len(images),
                "detected_objects": len(objects),
            }
        }

    def _extract_objects(self, outputs: dict, source_images: List[str]) -> List[dict]:
        """从重建结果中提取对象提案"""
        objects = []

        # 使用点云聚类检测对象
        point_cloud = outputs.get("fused_point_cloud", outputs.get("point_cloud"))

        if point_cloud is not None:
            # 简单的 DBSCAN 聚类
            from sklearn.cluster import DBSCAN
            clustering = DBSCAN(eps=0.3, min_samples=100).fit(point_cloud)

            for label in set(clustering.labels_):
                if label == -1:  # 噪声
                    continue

                mask = clustering.labels_ == label
                cluster_points = point_cloud[mask]

                # 计算包围盒
                bbox_min = cluster_points.min(axis=0)
                bbox_max = cluster_points.max(axis=0)
                center = (bbox_min + bbox_max) / 2

                objects.append({
                    "id": f"obj_{label:03d}",
                    "action": "create",
                    "object": {
                        "type": "other",  # 需要后续分类
                        "label": f"Object {label}",
                        "pose": {
                            "position": center.tolist(),
                            "rotation": [1, 0, 0, 0]
                        },
                        "bbox": {
                            "min": bbox_min.tolist(),
                            "max": bbox_max.tolist()
                        },
                        "state": "visible"
                    },
                    "confidence": 0.8,
                    "source_images": source_images
                })

        return objects

    def _save_pointcloud(self, outputs: dict) -> str:
        """保存点云文件"""
        point_cloud = outputs.get("fused_point_cloud", outputs.get("point_cloud"))
        if point_cloud is None:
            return None

        # 保存为 PLY 格式
        import uuid
        path = f"/tmp/pointcloud_{uuid.uuid4().hex}.ply"

        cloud = trimesh.PointCloud(point_cloud)
        cloud.export(path)

        return path

    def _save_mesh(self, outputs: dict) -> Optional[str]:
        """保存 mesh 文件"""
        if "mesh" not in outputs:
            return None

        import uuid
        path = f"/tmp/mesh_{uuid.uuid4().hex}.glb"

        mesh = outputs["mesh"]
        mesh.export(path)

        return path

    def _compute_uncertainty(self, outputs: dict) -> List[dict]:
        """计算不确定性区域"""
        uncertainty_regions = []

        # 基于深度图的方差计算不确定性
        if "depth" in outputs:
            depth = outputs["depth"]
            # 高方差区域 = 高不确定性
            # 简化实现：标记深度值接近最大值的区域
            max_depth = depth.max()
            uncertain_mask = depth > max_depth * 0.9

            if uncertain_mask.any():
                # 找到不确定区域的边界
                y_indices, x_indices = np.where(uncertain_mask)
                if len(y_indices) > 0:
                    uncertainty_regions.append({
                        "id": "unc_001",
                        "bbox": {
                            "min": [x_indices.min(), y_indices.min(), 0],
                            "max": [x_indices.max(), y_indices.max(), 0]
                        },
                        "level": "high",
                        "reason": "Depth estimation uncertain in this region"
                    })

        return uncertainty_regions
```

## Gradio Demo

```bash
# 启动本地 Demo
python app.py
```

访问 `http://localhost:7860` 使用交互界面。

## 3D Gaussian 优化

```bash
# 对推理结果进行 3D Gaussian 优化
python submodules/gsplat/examples/simple_trainer_worldmirror.py default \
    --data_factor 1 \
    --data_dir /path/to/inference_output \
    --result_dir /path/to/gs_output
```

## 性能优化

### 内存优化

```python
# 对于大图片，使用分块处理
model.config.tile_size = 512  # 分块大小
model.config.overlap = 64     # 重叠区域
```

### 批处理

```python
# 批量处理多张图片
images = [Image.open(p) for p in image_paths]
batch_outputs = model.predict_batch(images, batch_size=4)
```

## 降级策略

| 情况 | 降级方案 |
|------|----------|
| GPU 内存不足 | 降低输入分辨率或使用 tile 模式 |
| Mirror 不可用 | 切换到 HunyuanWorld-Voyager |
| 全部失败 | 返回 Mock SceneGraph + 高 uncertainty |

## 输出格式

### Point Cloud

```python
# shape: (N, 3) - N 个点，每个点有 x, y, z 坐标
point_cloud = outputs["point_cloud"]
print(f"Points: {len(point_cloud)}")
```

### Depth Map

```python
# shape: (H, W) - 与输入图像同尺寸
depth = outputs["depth"]
print(f"Depth range: {depth.min():.2f} - {depth.max():.2f} meters")
```

### Camera Parameters

```python
camera = outputs["camera"]
# intrinsics: (3, 3) 相机内参矩阵
# extrinsics: dict with "rotation" (3, 3) and "translation" (3,)
```

## 注意事项

1. **输入分辨率**：建议 720p-1080p，过高分辨率会显著增加内存占用
2. **多视角输入**：视角覆盖越全面，重建质量越高
3. **相机位姿**：如果有准确的相机位姿，作为先验输入会显著提升质量
4. **光照一致性**：输入图片应在相似光照条件下拍摄

## 参考链接

- [GitHub 仓库](https://github.com/Tencent-Hunyuan/HunyuanWorld-Mirror)
- [HuggingFace 模型](https://huggingface.co/tencent/HunyuanWorld-Mirror)
- [技术报告](https://arxiv.org/abs/...)
