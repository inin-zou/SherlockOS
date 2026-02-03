// Core domain types matching backend models

export interface Case {
  id: string;
  title: string;
  description?: string;
  created_at: string;
}

export interface Commit {
  id: string;
  case_id: string;
  parent_commit_id?: string;
  branch_id?: string;
  type: CommitType;
  summary: string;
  payload: Record<string, unknown>;
  created_at: string;
}

export type CommitType =
  | 'upload_scan'
  | 'witness_statement'
  | 'manual_edit'
  | 'reconstruction_update'
  | 'profile_update'
  | 'reasoning_result'
  | 'export_report'
  | 'replay_generated';

export interface Job {
  id: string;
  case_id: string;
  type: JobType;
  status: JobStatus;
  progress: number;
  input: Record<string, unknown>;
  output?: Record<string, unknown>;
  error?: string;
  created_at: string;
  updated_at: string;
}

export type JobType =
  | 'reconstruction'
  | 'imagegen'
  | 'reasoning'
  | 'profile'
  | 'export'
  | 'replay'
  | 'asset3d'
  | 'scene_analysis';

// Image generation types
export type ImageGenType =
  | 'portrait'
  | 'evidence_board'
  | 'comparison'
  | 'report_figure'
  | 'scene_pov'      // Generate consistent POV images for reconstruction
  | 'asset_clean';   // Generate clean isolated object for 3D

// POV generation input
export interface POVGenerationInput {
  case_id: string;
  gen_type: 'scene_pov';
  scene_description: string;
  view_angles: string[];  // ["front", "left", "right", "back", "corner_nw", "corner_se"]
  room_type?: string;
  resolution: '1k' | '2k' | '4k';
  reference_image_keys?: string[];
}

// Generated image in a batch
export interface GeneratedImage {
  view_angle: string;
  asset_key: string;
  thumbnail_key: string;
  width: number;
  height: number;
}

// Reconstruction input (hybrid mode supported)
export interface ReconstructionInput {
  case_id: string;
  scan_asset_keys: string[];           // Raw uploaded images (required)
  generated_pov_keys?: string[];       // Nano Banana generated POV images (optional)
  scene_description?: string;          // Scene description for POV generation
  enable_preprocess?: boolean;         // If true, generate POV images first
  room_type?: string;                  // "office", "bedroom", etc.
}

export type JobStatus = 'queued' | 'running' | 'done' | 'failed' | 'canceled';

export interface PointCloud {
  positions: number[][]; // [[x,y,z], ...]
  colors?: number[][];   // [[r,g,b], ...] in 0-1 range
  count: number;
}

export interface SceneGraph {
  version: string;
  bounds: BoundingBox;
  objects: SceneObject[];
  evidence: EvidenceCard[];
  constraints: Constraint[];
  uncertainty_regions?: UncertaintyRegion[];
  point_cloud?: PointCloud;
}

export interface SceneObject {
  id: string;
  type: ObjectType;
  label: string;
  pose: Pose;
  bbox: BoundingBox;
  mesh_ref?: string;
  state: ObjectState;
  evidence_ids: string[];
  confidence: number;
  source_commit_ids: string[];
  metadata?: Record<string, unknown>;
}

export type ObjectType =
  | 'furniture'
  | 'door'
  | 'window'
  | 'wall'
  | 'evidence_item'
  | 'weapon'
  | 'footprint'
  | 'bloodstain'
  | 'vehicle'
  | 'person_marker'
  | 'other';

export type ObjectState = 'visible' | 'occluded' | 'suspicious' | 'removed';

export interface Pose {
  position: [number, number, number];
  rotation: [number, number, number, number];
  scale?: [number, number, number];
}

export interface BoundingBox {
  min: [number, number, number];
  max: [number, number, number];
}

export interface EvidenceCard {
  id: string;
  object_ids: string[];
  title: string;
  description: string;
  confidence: number;
  sources: EvidenceSource[];
  conflicts?: EvidenceSource[];
  created_at: string;
}

export interface EvidenceSource {
  type: 'upload' | 'witness' | 'inference';
  commit_id: string;
  description?: string;
  credibility?: number;
}

export interface Constraint {
  id: string;
  type: ConstraintType;
  description: string;
  params: Record<string, unknown>;
  confidence: number;
}

export type ConstraintType =
  | 'door_direction'
  | 'passable_area'
  | 'height_range'
  | 'time_window'
  | 'custom';

export interface UncertaintyRegion {
  id: string;
  bbox: BoundingBox;
  level: 'low' | 'medium' | 'high';
  reason: string;
}

export interface Trajectory {
  id: string;
  rank: number;
  overall_confidence: number;
  segments: TrajectorySegment[];
}

export interface TrajectorySegment {
  id: string;
  from_position: [number, number, number];
  to_position: [number, number, number];
  waypoints?: [number, number, number][];
  time_estimate?: { start: string; end: string };
  evidence_refs: EvidenceRef[];
  confidence: number;
  explanation: string;
}

export interface EvidenceRef {
  evidence_id: string;
  object_id?: string;
  relevance: 'supports' | 'contradicts' | 'neutral';
  weight: number;
}

export interface SuspectProfile {
  case_id: string;
  commit_id: string;
  attributes: SuspectAttributes;
  portrait_asset_key?: string;
  updated_at: string;
}

export interface SuspectAttributes {
  age_range?: { min: number; max: number; confidence: number };
  height_range_cm?: { min: number; max: number; confidence: number };
  build?: { value: 'slim' | 'average' | 'heavy'; confidence: number };
  skin_tone?: { value: string; confidence: number };
  hair?: { style: string; color: string; confidence: number };
  facial_hair?: { type: string; confidence: number };
  glasses?: { type: string; confidence: number };
  distinctive_features?: Array<{ description: string; confidence: number }>;
}

export interface Asset {
  id: string;
  case_id: string;
  kind: AssetKind;
  storage_key: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export type AssetKind =
  | 'scan_image'
  | 'generated_image'
  | 'mesh'
  | 'pointcloud'
  | 'portrait'
  | 'report'
  | 'replay_video'
  | 'evidence_model';

// Evidence Archive types for sidebar
export interface EvidenceFolder {
  id: string;
  name: string;
  icon: string;
  items: EvidenceItem[];
  isOpen?: boolean;
}

export interface EvidenceItem {
  id: string;
  name: string;
  type: 'pdf' | 'image' | 'video' | 'audio' | 'json' | 'text' | '3d';
  assetKey?: string;
  metadata?: Record<string, unknown>;
}

// Timeline track types
export interface TimelineTrack {
  id: string;
  name: string;
  type: 'location' | 'person' | 'event';
  color: string;
  events: TimelineEvent[];
}

export interface TimelineEvent {
  id: string;
  start: number; // timestamp or frame number
  end: number;
  label?: string;
  confidence?: number;
  relatedObjects?: string[];
}

// 3D Scene annotation
export interface SceneAnnotation {
  id: string;
  position: [number, number, number];
  label: string;
  type: 'marker' | 'label' | 'path';
  color?: string;
  visible?: boolean;
}

// API Response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
  meta?: {
    cursor?: string;
    total?: number;
  };
}
