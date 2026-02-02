-- SherlockOS Database Schema Update
-- Migration: 003_add_new_job_and_asset_types
-- Description: Add new job types and asset kinds for expanded model architecture
--   - replay: HY-World-1.5 trajectory replay animation
--   - asset3d: Hunyuan3D-2 evidence 3D model generation
--   - scene_analysis: Gemini 3 Pro Preview scene understanding
--   - replay_video: Video output from HY-World-1.5
--   - evidence_model: GLB model output from Hunyuan3D-2

-- ============================================
-- ADD NEW JOB TYPES
-- ============================================

-- Add 'replay' job type for HY-World-1.5 trajectory animation
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'replay';

-- Add 'asset3d' job type for Hunyuan3D-2 evidence model generation
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'asset3d';

-- Add 'scene_analysis' job type for Gemini 3 Pro Preview vision
ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'scene_analysis';

-- ============================================
-- ADD NEW ASSET KINDS
-- ============================================

-- Add 'replay_video' asset kind for HY-World-1.5 output (MP4)
ALTER TYPE asset_kind ADD VALUE IF NOT EXISTS 'replay_video';

-- Add 'evidence_model' asset kind for Hunyuan3D-2 output (GLB)
ALTER TYPE asset_kind ADD VALUE IF NOT EXISTS 'evidence_model';

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TYPE job_type IS 'Async job types:
  - reconstruction: HunyuanWorld-Mirror 3D scene reconstruction
  - imagegen: Nano Banana image generation
  - reasoning: Gemini 2.5 Flash trajectory reasoning
  - profile: Gemini 2.5 Flash suspect profile extraction
  - export: HTML/PDF report generation
  - replay: HY-World-1.5 trajectory video animation
  - asset3d: Hunyuan3D-2 evidence 3D model generation
  - scene_analysis: Gemini 3 Pro Preview scene understanding';

COMMENT ON TYPE asset_kind IS 'Asset types:
  - scan_image: Uploaded crime scene scan images
  - generated_image: Nano Banana generated images
  - mesh: 3D mesh files (GLB/OBJ)
  - pointcloud: Point cloud data
  - portrait: Suspect portrait images
  - report: Generated HTML/PDF reports
  - replay_video: HY-World-1.5 trajectory animation videos
  - evidence_model: Hunyuan3D-2 evidence 3D models (GLB)';
