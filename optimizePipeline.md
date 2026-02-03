# SherlockOS Optimized AI Pipeline

> **Version:** 1.0
> **Last Updated:** 2026-02-03
> **Status:** Implementation Plan

---

## Problem Statement

### What We Encountered

After implementing the initial reconstruction pipeline using HunyuanWorld-Mirror, we ran into several issues:

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ORIGINAL PIPELINE PROBLEMS                                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Problem 1: Point Cloud Doesn't Look Like a Room                        │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Mirror produced a point cloud, but it was object-centric │       │
│  │  • The output was tiny/floating, not room-scale             │       │
│  │  • Colors were washed out (gray/white)                      │       │
│  │  • Didn't resemble the actual crime scene                   │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
│  Problem 2: Inconsistent Input Images                                   │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Raw crime scene photos have varying lighting             │       │
│  │  • Different angles, exposures, color temperatures          │       │
│  │  • Some areas over/under exposed                            │       │
│  │  • Mirror struggles with inconsistent multi-view input      │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
│  Problem 3: Coverage Gaps                                               │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Photos don't cover all angles of the room                │       │
│  │  • Missing viewpoints = holes in reconstruction             │       │
│  │  • 3D models need 360° coverage for quality output          │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Root Cause Analysis

| Issue | Root Cause | Impact |
|-------|------------|--------|
| Tiny point cloud | Mirror outputs in normalized coordinates (~[-1,1]) | Required manual scaling to fit room |
| Washed out colors | Source images were dark (avg 17% brightness) | Added brightness/saturation boost |
| Doesn't look like room | Mirror is optimized for objects, not room-scale scenes | Fundamental model limitation |
| Inconsistent quality | Raw photos have varying conditions | Model gets confused by lighting differences |

### Solutions We Tried (Partial Success)

1. **Scaling**: Scaled point cloud to fit room bounds → helped with size
2. **Color boost**: 1.8x brightness, 1.3x saturation → helped with visibility
3. **Remove room walls**: Show only point cloud → looked worse without context
4. **Proxy Geometry**: Keep room shell, overlay point cloud → better but still not great

### The Real Solution: Fix the Input, Not the Output

Instead of trying to fix Mirror's output, we should **improve the input**:

```
Before: Raw inconsistent photos → Mirror → Poor reconstruction
After:  Raw photos + AI-generated consistent POVs → Mirror → Better reconstruction
```

---

## Why This Approach

### Key Insight

**Nano Banana (Gemini Image Gen) can normalize inputs** - generating consistent, well-lit, properly-angled views of the same scene that Mirror can reconstruct more effectively.

### Benefits of Hybrid Input

| Benefit | Explanation |
|---------|-------------|
| **Forensic accuracy preserved** | Raw photos kept as Tier 1 evidence |
| **Consistent lighting** | Generated POVs have uniform lighting |
| **Complete coverage** | Generate views for missing angles |
| **Better reconstruction** | Mirror gets cleaner multi-view input |
| **Best of both worlds** | Accuracy from raw + quality from generated |

### Why Not Just Use Generated Images?

For a **crime scene investigation tool**, we must preserve forensic accuracy:

- Raw photos = **Ground Truth** (Tier 1) - what was actually there
- Generated photos = **Visualization Aid** - helps AI understand the space
- Both together = Accurate details + good 3D reconstruction

---

## Key Decisions Made

During our troubleshooting discussion, we made the following architectural decisions:

### Decision 1: Use Proxy Geometry for Visualization
> **Problem**: Point cloud alone doesn't provide spatial context
> **Decision**: Always show room shell (walls/floor/ceiling) as Tier 0 boundary, overlay point cloud as "scan data"
> **Rationale**: Provides believable environment while showing what AI detected

### Decision 2: Nano Banana as Preprocessor for Reconstruction
> **Problem**: Raw photos are inconsistent, leading to poor Mirror output
> **Decision**: Generate consistent POV images with Nano Banana, combine with raw photos
> **Rationale**: Raw photos preserve accuracy, generated images improve coverage and consistency

### Decision 3: Pass Both Raw + Generated Images to Mirror
> **Problem**: Using only generated images loses forensic details
> **Decision**: Feed combined set (raw + generated) to Mirror together
> **Rationale**: More input data = better reconstruction; raw preserves truth, generated fills gaps

### Decision 4: Two-Stage 3D Asset Generation
> **Problem**: Evidence items in crime scene photos have clutter/context
> **Decision**: Nano Banana generates clean isolated image → Hunyuan3D-2.1 generates 3D model
> **Rationale**: Clean input = higher quality 3D output

### Decision 5: Mirror Image Limit ~8-12
> **Problem**: How many images can Mirror handle?
> **Decision**: Optimal range is 8-12 images (A100-80GB can handle this)
> **Rationale**: Enough for good coverage without memory issues; diminishing returns after ~10

### Decision 6: Why Not Camera Priors? (Future-Proofing)
> **Problem**: Mirror demos look great because they use camera poses, intrinsics, depth maps
> **Our Situation**: We only have random photos, no camera metadata
> **Decision**: Use Nano Banana preprocessing instead of requiring camera priors
> **Rationale**:
> - Works with current input (random photos)
> - Same pipeline works later when we have video with camera data
> - No complex preprocessing (COLMAP) required now
> - Future-proof: better input = better output, no pipeline changes needed

```
NOW:    Raw Photos + Nano Banana POVs ──→ Mirror ──→ Decent Result
LATER:  Video Frames + Camera Priors ──→ Mirror ──→ Great Result
        (Same pipeline, just better input)
```

---

## Overview

This document outlines the optimized AI pipeline for SherlockOS, focusing on maximizing reconstruction quality and evidence visualization through strategic use of multiple Hunyuan models with Nano Banana as a preprocessing layer.

### Core Principle

**Nano Banana (Gemini Image Gen) serves as the "normalizer"** - preprocessing raw inputs into consistent, high-quality images that downstream models can process more effectively.

---

## Model Ecosystem

| Model | Purpose | Input | Output |
|-------|---------|-------|--------|
| **Nano Banana** | Image generation/normalization | Text prompt + reference image | Clean, consistent images |
| **HunyuanWorld-Mirror** | 3D scene reconstruction | Multi-view images (8-12) | Point cloud + depth maps |
| **Hunyuan3D-2.1** | Evidence 3D asset generation | Single clean object image | 3D mesh (.glb/.obj) |
| **HY-WorldPlay** | Trajectory replay video | Scene + camera trajectory | Video file |
| **Gemini 2.5 Flash** | Scene analysis & reasoning | Images + SceneGraph | Detected objects, trajectories |

---

## Pipeline 1: Hybrid Scene Reconstruction

### Problem
Raw crime scene photos have inconsistent lighting, angles, exposure, and color temperature. This leads to poor 3D reconstruction quality.

### Solution
Combine raw photos (forensic accuracy) with Nano Banana generated images (consistent coverage) as input to Mirror.

```
┌─────────────────────────────────────────────────────────────────────────┐
│  HYBRID RECONSTRUCTION PIPELINE                                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Step 1: Raw Photo Upload                                               │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  User uploads crime scene photos (4-6 images)               │       │
│  │  ├── 01_scene_overview.png                                  │       │
│  │  ├── 02_entry_point.png                                     │       │
│  │  ├── 03_desk_area.png                                       │       │
│  │  ├── 04_evidence_closeup.png                                │       │
│  │  └── 05_exit_door.png                                       │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 2: Scene Analysis (Gemini)                                        │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Analyze photos to understand:                              │       │
│  │  • Room type and layout                                     │       │
│  │  • Key objects and their positions                          │       │
│  │  • Evidence items detected                                  │       │
│  │  • Scene description for Nano Banana prompts                │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 3: Nano Banana - Generate Consistent POV Images                   │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Prompt: "Crime scene office interior, forensic             │       │
│  │          documentation style, consistent lighting,          │       │
│  │          POV: {front/left/right/back/corner}, showing       │       │
│  │          {scene_description from Step 2}"                   │       │
│  │                                                             │       │
│  │  Generated Images (4-6):                                    │       │
│  │  ├── gen_front_view.png                                     │       │
│  │  ├── gen_left_view.png                                      │       │
│  │  ├── gen_right_view.png                                     │       │
│  │  ├── gen_back_view.png                                      │       │
│  │  ├── gen_corner_nw.png                                      │       │
│  │  └── gen_corner_se.png                                      │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 4: Combined Input to Mirror                                       │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Raw Photos (5)        +    Generated POVs (6)              │       │
│  │  ───────────────────────────────────────────────            │       │
│  │  = 11 images total (optimal range: 8-12)                    │       │
│  │                                                             │       │
│  │  Benefits:                                                  │       │
│  │  • Raw photos preserve forensic details                     │       │
│  │  • Generated POVs fill coverage gaps                        │       │
│  │  • Consistent lighting aids reconstruction                  │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 5: HunyuanWorld-Mirror                                            │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Output:                                                    │       │
│  │  • High-quality point cloud (up to 50k points)              │       │
│  │  • Depth maps for each view                                 │       │
│  │  • Camera pose estimates                                    │       │
│  │  • Uncertainty regions                                      │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Implementation

```typescript
// ReconstructionInput (enhanced)
interface ReconstructionInput {
  case_id: string;
  raw_image_keys: string[];      // Original crime scene photos (Tier 1)
  gen_image_keys?: string[];     // Nano Banana generated POVs (optional)
  scene_description?: string;    // From scene analysis, for POV generation
  skip_preprocessing?: boolean;  // Use raw images only
}
```

### Job Flow

```
1. upload_scan → stores raw images
2. scene_analysis (Gemini) → detects objects, generates scene description
3. imagegen (Nano Banana) → generates consistent POV images
4. reconstruction (Mirror) → receives raw + generated images
5. Result: Better quality 3D reconstruction
```

---

## Pipeline 2: Evidence 3D Asset Generation

### Problem
Crime scene photos show evidence items in context (on tables, floors, with other objects). Hunyuan3D-2.1 produces better 3D models from clean, isolated images.

### Solution
Use Nano Banana to generate a clean, isolated image of the evidence item, then feed to Hunyuan3D-2.1.

```
┌─────────────────────────────────────────────────────────────────────────┐
│  EVIDENCE 3D ASSET PIPELINE                                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Step 1: Scene Analysis Detection                                       │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Gemini detects evidence in crime scene photo:              │       │
│  │  {                                                          │       │
│  │    "id": "evidence_001",                                    │       │
│  │    "type": "weapon",                                        │       │
│  │    "label": "Kitchen knife",                                │       │
│  │    "description": "8-inch chef's knife with black handle",  │       │
│  │    "bounding_box": [x1, y1, x2, y2],                        │       │
│  │    "source_image_key": "03_desk_area.png"                   │       │
│  │  }                                                          │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 2: Nano Banana - Generate Clean Asset Image                       │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Input:                                                     │       │
│  │  • Reference: Cropped region from source image              │       │
│  │  • Prompt: "Forensic evidence photo of 8-inch chef's knife  │       │
│  │            with black handle, isolated on white background, │       │
│  │            studio lighting, high detail, multiple angles"   │       │
│  │                                                             │       │
│  │  Output: Clean isolated object image (1024x1024)            │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 3: Hunyuan3D-2.1 - Generate 3D Model                              │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Input: Clean isolated image from Step 2                    │       │
│  │                                                             │       │
│  │  Output:                                                    │       │
│  │  • 3D mesh (.glb format)                                    │       │
│  │  • Texture maps                                             │       │
│  │  • Normal maps                                              │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 4: Place in Scene                                                 │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Load 3D model at detected position                       │       │
│  │  • User can rotate/inspect from any angle                   │       │
│  │  • Link to evidence card for investigation                  │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Implementation

```typescript
// Asset3DInput
interface Asset3DInput {
  case_id: string;
  evidence_id: string;           // Reference to detected evidence
  object_description: string;    // From scene analysis
  reference_image_key: string;   // Original image containing object
  bounding_box?: number[];       // [x1, y1, x2, y2] crop region
}

// Asset3DOutput
interface Asset3DOutput {
  mesh_asset_key: string;        // Storage key for .glb file
  texture_asset_key?: string;    // Storage key for textures
  thumbnail_key: string;         // Preview image
  vertex_count: number;
  generation_time_ms: number;
}
```

### Job Flow

```
1. scene_analysis → detects "knife" evidence
2. User clicks "Generate 3D Model" on evidence card
3. imagegen (Nano Banana) → generates clean isolated image
4. asset3d (Hunyuan3D-2.1) → generates 3D mesh
5. Result: Inspectable 3D model in scene
```

---

## Pipeline 3: Trajectory Replay Video (WorldPlay)

### Purpose
After reasoning generates suspect movement trajectories, WorldPlay creates a video showing the path through the scene.

```
┌─────────────────────────────────────────────────────────────────────────┐
│  TRAJECTORY REPLAY PIPELINE                                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Prerequisites:                                                         │
│  • Scene reconstruction complete (point cloud available)                │
│  • Reasoning job complete (trajectories generated)                      │
│                                                                         │
│  Step 1: User Selects Trajectory                                        │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Trajectory #1 (87% confidence)                             │       │
│  │  "Suspect entered via window, moved to desk,                │       │
│  │   accessed filing cabinet, exited through back door"        │       │
│  │                                                             │       │
│  │  [Generate Replay Video]  [First Person]  [Third Person]    │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 2: Build Camera Trajectory                                        │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Convert trajectory segments to camera poses:               │       │
│  │                                                             │       │
│  │  Segment 1: Window → Desk                                   │       │
│  │    camera_poses: "w-20,right-15,forward-30"                 │       │
│  │                                                             │       │
│  │  Segment 2: Desk → Filing Cabinet                           │       │
│  │    camera_poses: "left-45,forward-20"                       │       │
│  │                                                             │       │
│  │  Segment 3: Filing Cabinet → Exit                           │       │
│  │    camera_poses: "left-90,forward-40"                       │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 3: HY-WorldPlay - Generate Video                                  │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  Input:                                                     │       │
│  │  • Reference image (scene overview)                         │       │
│  │  • Scene description                                        │       │
│  │  • Camera trajectory commands                               │       │
│  │  • Perspective: first_person / third_person                 │       │
│  │  • Frame count: 125 (5 seconds @ 24fps)                     │       │
│  │  • Resolution: 720p                                         │       │
│  │                                                             │       │
│  │  Output:                                                    │       │
│  │  • MP4 video file                                           │       │
│  │  • Thumbnail                                                │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                              │                                          │
│                              ▼                                          │
│  Step 4: Display in UI                                                  │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Video player in timeline panel                           │       │
│  │  • Synced with 3D view (camera follows video)               │       │
│  │  • Evidence highlights at key moments                       │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Implementation

```typescript
// ReplayInput (existing, enhanced)
interface ReplayInput {
  case_id: string;
  trajectory_id: string;
  trajectory?: Trajectory;
  reference_image_key: string;     // Scene overview image
  scene_description: string;       // From scene analysis
  camera_poses: string;            // Camera movement commands
  perspective: 'first_person' | 'third_person';
  frame_count: number;             // Default: 125
  resolution: '480p' | '720p';
}
```

---

## Visualization: Proxy Geometry Approach

### Principle
Per the SherlockOS specs, the frontend uses **Proxy Geometry** - simple geometric shapes representing the scene, with point cloud overlaid as "scanned data".

```
┌─────────────────────────────────────────────────────────────────────────┐
│  FRONTEND VISUALIZATION LAYERS                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Layer 1: Room Shell (Tier 0 - Physical Boundaries)                     │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Walls, floor, ceiling                                    │       │
│  │  • Doors, windows                                           │       │
│  │  • Always visible as base environment                       │       │
│  │  • Represents "immutable physics"                           │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
│  Layer 2: Proxy Geometry Objects                                        │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Furniture as boxes/cylinders                             │       │
│  │  • Evidence markers with labels                             │       │
│  │  • Person silhouettes                                       │       │
│  │  • Based on scene analysis detection                        │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
│  Layer 3: Point Cloud Overlay (Scan Data)                               │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Semi-transparent point cloud from Mirror                 │       │
│  │  • Shows "what the AI sees"                                 │       │
│  │  • Can be toggled on/off                                    │       │
│  │  • Labeled as "SCAN DATA"                                   │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
│  Layer 4: 3D Evidence Models (Optional)                                 │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • High-detail 3D models from Hunyuan3D-2.1                 │       │
│  │  • Only for key evidence items                              │       │
│  │  • User can inspect from any angle                          │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
│  Layer 5: Trajectory Visualization                                      │
│  ┌─────────────────────────────────────────────────────────────┐       │
│  │  • Animated paths from reasoning                            │       │
│  │  • Annotations at key locations                             │       │
│  │  • Discrepancy highlights                                   │       │
│  └─────────────────────────────────────────────────────────────┘       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Job Dependencies & Flow

```
                    ┌──────────────────┐
                    │   Upload Scan    │
                    │  (raw photos)    │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  Scene Analysis  │
                    │    (Gemini)      │
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
     ┌────────────────┐ ┌────────────┐ ┌────────────────┐
     │ ImageGen: POVs │ │ Reasoning  │ │ ImageGen: Asset│
     │ (Nano Banana)  │ │  (Gemini)  │ │ (Nano Banana)  │
     └───────┬────────┘ └─────┬──────┘ └───────┬────────┘
             │                │                │
             ▼                │                ▼
     ┌────────────────┐       │        ┌────────────────┐
     │ Reconstruction │       │        │    Asset3D     │
     │    (Mirror)    │       │        │ (Hunyuan3D-2.1)│
     └───────┬────────┘       │        └───────┬────────┘
             │                │                │
             └────────────────┼────────────────┘
                              │
                              ▼
                     ┌────────────────┐
                     │     Replay     │
                     │  (WorldPlay)   │
                     └────────────────┘
```

---

## API Changes Summary

### New/Modified Endpoints

| Endpoint | Change | Description |
|----------|--------|-------------|
| `POST /v1/cases/{id}/jobs` | Enhanced | Support `reconstruction` with `gen_image_keys` |
| `POST /v1/cases/{id}/jobs` | New type | `asset3d` job type for 3D evidence models |
| `POST /v1/cases/{id}/jobs` | Enhanced | `imagegen` supports `purpose: "pov_generation"` |

### New Job Types

```typescript
type JobType =
  | 'reconstruction'    // Enhanced: raw + generated images
  | 'scene_analysis'    // Unchanged
  | 'imagegen'          // Enhanced: POV generation mode
  | 'asset3d'           // NEW: Evidence 3D model generation
  | 'reasoning'         // Unchanged
  | 'replay'            // Unchanged
  | 'export';           // Unchanged
```

---

## Implementation Priority

### Phase 1: Enhanced Reconstruction (High Priority)
1. [x] Modify `imagegen` worker to support POV generation mode (`scene_pov` gen_type)
2. [x] Update `reconstruction` worker to accept combined image sets (`generated_pov_keys`)
3. [x] Add Supabase Storage upload for generated POV images
4. [x] Add preprocessing step to reconstruction job flow (auto-trigger POV gen via `enable_preprocess`)
5. [ ] Test full hybrid pipeline with demo images

### Phase 2: Evidence 3D Assets (Medium Priority)
1. [x] Create Replicate endpoint for Hunyuan3D-2.1
2. [x] Implement `asset3d` worker
3. [x] Implement `asset_clean` image generation type
4. [ ] Add "Generate 3D Model" button to evidence cards
5. [ ] Integrate 3D model loading in SceneViewer

### Phase 3: WorldPlay Integration (Lower Priority)
1. [x] Create Modal endpoint for HY-WorldPlay
2. [x] Implement `replay` worker
3. [ ] Implement camera trajectory builder (trajectory → camera poses)
4. [ ] Add video player to timeline panel
5. [ ] Sync video playback with 3D view

---

## Resource Estimates

| Model | GPU | Memory | Typical Time |
|-------|-----|--------|--------------|
| Nano Banana | - | - | 5-10s per image |
| HunyuanWorld-Mirror | A100-80GB | ~40GB | 30-60s |
| Hunyuan3D-2.1 | A100-40GB | ~20GB | 20-40s |
| HY-WorldPlay | A100-80GB | ~60GB | 60-120s |

### Cost Optimization
- Batch POV generation (all 6 images in one call if possible)
- Cache generated POVs for re-reconstruction
- Only generate 3D models on user request (not automatic)

---

## Data Flow & Storage

### Tier Classification

| Tier | Data Type | Storage | Mutability |
|------|-----------|---------|------------|
| **Tier 0** | Room boundaries, physics | SceneGraph | Immutable |
| **Tier 1** | Raw photos (evidence) | Supabase Storage | Immutable |
| **Tier 1.5** | AI-generated POVs | Supabase Storage | Regeneratable |
| **Tier 2** | Point cloud, 3D models | Supabase Storage | Regeneratable |
| **Tier 3** | Trajectories, reasoning | SceneGraph | Updatable |

### Asset Keys Convention

```
cases/{case_id}/raw/{batch_id}/{filename}        # Original uploads
cases/{case_id}/generated/pov/{filename}         # Nano Banana POVs
cases/{case_id}/generated/asset/{evidence_id}/   # Clean asset images
cases/{case_id}/models/{evidence_id}.glb         # 3D models
cases/{case_id}/replay/{trajectory_id}.mp4       # Replay videos
```

---

*Document Version: 1.0 | Created: 2026-02-03*
