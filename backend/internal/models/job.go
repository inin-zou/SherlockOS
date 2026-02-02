package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Job represents an async processing job
type Job struct {
	ID             uuid.UUID       `json:"id"`
	CaseID         uuid.UUID       `json:"case_id"`
	Type           JobType         `json:"type"`
	Status         JobStatus       `json:"status"`
	Progress       int             `json:"progress"`
	Input          json.RawMessage `json:"input"`
	Output         json.RawMessage `json:"output,omitempty"`
	Error          string          `json:"error,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	RetryCount     int             `json:"retry_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// Validate checks if the Job is valid
func (j *Job) Validate() error {
	if j.CaseID == uuid.Nil {
		return errors.New("case_id is required")
	}
	if !j.Type.IsValid() {
		return errors.New("invalid job type")
	}
	if !j.Status.IsValid() {
		return errors.New("invalid job status")
	}
	if j.Progress < 0 || j.Progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}
	return nil
}

// NewJob creates a new Job with generated ID and timestamps
func NewJob(caseID uuid.UUID, jobType JobType, input interface{}) (*Job, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Job{
		ID:         uuid.New(),
		CaseID:     caseID,
		Type:       jobType,
		Status:     JobStatusQueued,
		Progress:   0,
		Input:      inputBytes,
		RetryCount: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// SetIdempotencyKey sets the idempotency key for the job
func (j *Job) SetIdempotencyKey(key string) {
	j.IdempotencyKey = key
}

// MarkRunning transitions the job to running status
func (j *Job) MarkRunning() error {
	if j.Status != JobStatusQueued {
		return errors.New("job must be in queued status to start running")
	}
	j.Status = JobStatusRunning
	j.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateProgress updates the job progress
func (j *Job) UpdateProgress(progress int) error {
	if progress < 0 || progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}
	j.Progress = progress
	j.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkDone transitions the job to done status with output
func (j *Job) MarkDone(output interface{}) error {
	if j.Status != JobStatusRunning {
		return errors.New("job must be in running status to complete")
	}
	outputBytes, err := json.Marshal(output)
	if err != nil {
		return err
	}
	j.Status = JobStatusDone
	j.Progress = 100
	j.Output = outputBytes
	j.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkFailed transitions the job to failed status with error message
func (j *Job) MarkFailed(errMsg string) {
	j.Status = JobStatusFailed
	j.Error = errMsg
	j.UpdatedAt = time.Now().UTC()
}

// MarkCanceled transitions the job to canceled status
func (j *Job) MarkCanceled() {
	j.Status = JobStatusCanceled
	j.UpdatedAt = time.Now().UTC()
}

// IncrementRetry increments the retry count
func (j *Job) IncrementRetry() {
	j.RetryCount++
	j.UpdatedAt = time.Now().UTC()
}

// Heartbeat updates the timestamp to indicate the job is still alive
func (j *Job) Heartbeat() {
	j.UpdatedAt = time.Now().UTC()
}

// ReconstructionInput represents input for reconstruction jobs
type ReconstructionInput struct {
	CaseID             string       `json:"case_id"`
	ScanAssetKeys      []string     `json:"scan_asset_keys"`
	CameraPoses        []CameraPose `json:"camera_poses,omitempty"`
	DepthMaps          []string     `json:"depth_maps,omitempty"`
	ExistingScenegraph *SceneGraph  `json:"existing_scenegraph,omitempty"`
}

// Validate checks if the ReconstructionInput is valid
func (r *ReconstructionInput) Validate() error {
	if r.CaseID == "" {
		return errors.New("case_id is required")
	}
	if len(r.ScanAssetKeys) == 0 {
		return errors.New("at least one scan_asset_key is required")
	}
	for i, key := range r.ScanAssetKeys {
		if key == "" {
			return errors.New("scan_asset_key at index " + string(rune('0'+i)) + " is empty")
		}
	}
	return nil
}

// CameraPose represents camera intrinsics and extrinsics
type CameraPose struct {
	AssetKey   string         `json:"asset_key"`
	Intrinsics CameraIntrinsics `json:"intrinsics"`
	Extrinsics CameraExtrinsics `json:"extrinsics"`
}

// CameraIntrinsics represents camera intrinsic parameters
type CameraIntrinsics struct {
	Fx float64 `json:"fx"`
	Fy float64 `json:"fy"`
	Cx float64 `json:"cx"`
	Cy float64 `json:"cy"`
}

// CameraExtrinsics represents camera extrinsic parameters
type CameraExtrinsics struct {
	Rotation    []float64 `json:"rotation"`
	Translation []float64 `json:"translation"`
}

// ReconstructionOutput represents output from reconstruction jobs
type ReconstructionOutput struct {
	Objects             []SceneObjectProposal `json:"objects"`
	MeshAssetKey        string                `json:"mesh_asset_key,omitempty"`
	PointcloudAssetKey  string                `json:"pointcloud_asset_key,omitempty"`
	UncertaintyRegions  []UncertaintyRegion   `json:"uncertainty_regions"`
	ProcessingStats     ProcessingStats       `json:"processing_stats"`
}

// SceneObjectProposal represents a proposed change to a scene object
type SceneObjectProposal struct {
	ID           string       `json:"id"`
	Action       string       `json:"action"` // "create", "update", "remove"
	Object       *SceneObject `json:"object,omitempty"`
	Confidence   float64      `json:"confidence"`
	SourceImages []string     `json:"source_images"`
}

// ProcessingStats contains statistics about processing
type ProcessingStats struct {
	InputImages      int   `json:"input_images"`
	DetectedObjects  int   `json:"detected_objects"`
	ProcessingTimeMs int64 `json:"processing_time_ms"`
}

// ReasoningInput represents input for reasoning jobs
type ReasoningInput struct {
	CaseID              string       `json:"case_id"`
	Scenegraph          *SceneGraph  `json:"scenegraph"`
	BranchID            string       `json:"branch_id,omitempty"`
	ConstraintsOverride []Constraint `json:"constraints_override,omitempty"`
	ThinkingBudget      int          `json:"thinking_budget,omitempty"`
	MaxTrajectories     int          `json:"max_trajectories,omitempty"`
}

// Validate checks if the ReasoningInput is valid
func (r *ReasoningInput) Validate() error {
	if r.CaseID == "" {
		return errors.New("case_id is required")
	}
	if r.Scenegraph == nil {
		return errors.New("scenegraph is required")
	}
	if r.ThinkingBudget < 0 || r.ThinkingBudget > 24576 {
		return errors.New("thinking_budget must be between 0 and 24576")
	}
	if r.MaxTrajectories < 0 {
		return errors.New("max_trajectories must be non-negative")
	}
	return nil
}

// SetDefaults sets default values for ReasoningInput
func (r *ReasoningInput) SetDefaults() {
	if r.ThinkingBudget == 0 {
		r.ThinkingBudget = 8192 // Default thinking budget
	}
	if r.MaxTrajectories == 0 {
		r.MaxTrajectories = 3 // Default top-K trajectories
	}
}

// ReasoningOutput represents output from reasoning jobs
type ReasoningOutput struct {
	Trajectories        []Trajectory        `json:"trajectories"`
	UncertaintyAreas    []UncertaintyRegion `json:"uncertainty_areas"`
	NextStepSuggestions []Suggestion        `json:"next_step_suggestions"`
	ThinkingSummary     string              `json:"thinking_summary,omitempty"`
	ModelStats          ModelStats          `json:"model_stats"`
}

// Trajectory represents a possible movement path
type Trajectory struct {
	ID                string             `json:"id"`
	Rank              int                `json:"rank"`
	OverallConfidence float64            `json:"overall_confidence"`
	Segments          []TrajectorySegment `json:"segments"`
}

// TrajectorySegment represents a segment of a trajectory
type TrajectorySegment struct {
	ID           string         `json:"id"`
	FromPosition [3]float64     `json:"from_position"`
	ToPosition   [3]float64     `json:"to_position"`
	Waypoints    [][3]float64   `json:"waypoints,omitempty"`
	TimeEstimate *TimeEstimate  `json:"time_estimate,omitempty"`
	EvidenceRefs []EvidenceRef  `json:"evidence_refs"`
	Confidence   float64        `json:"confidence"`
	Explanation  string         `json:"explanation"`
}

// TimeEstimate represents a time window estimate
type TimeEstimate struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// EvidenceRef represents a reference to evidence
type EvidenceRef struct {
	EvidenceID string  `json:"evidence_id"`
	ObjectID   string  `json:"object_id,omitempty"`
	Relevance  string  `json:"relevance"` // "supports", "contradicts", "neutral"
	Weight     float64 `json:"weight"`
}

// Suggestion represents a next step suggestion
type Suggestion struct {
	Type             string   `json:"type"` // "collect_evidence", "verify_constraint", "interview", "analyze"
	Description      string   `json:"description"`
	Priority         string   `json:"priority"` // "high", "medium", "low"
	RelatedObjectIDs []string `json:"related_object_ids,omitempty"`
}

// ModelStats contains statistics about model usage
type ModelStats struct {
	ThinkingTokens int   `json:"thinking_tokens"`
	OutputTokens   int   `json:"output_tokens"`
	LatencyMs      int64 `json:"latency_ms"`
}

// ImageGenInput represents input for image generation jobs
type ImageGenInput struct {
	CaseID            string             `json:"case_id"`
	GenType           ImageGenType       `json:"gen_type"`
	PortraitAttrs     *SuspectAttributes `json:"portrait_attributes,omitempty"`
	ReferenceImageKey string             `json:"reference_image_key,omitempty"`
	ObjectIDs         []string           `json:"object_ids,omitempty"`
	Layout            string             `json:"layout,omitempty"`
	Resolution        string             `json:"resolution"`
	StylePrompt       string             `json:"style_prompt,omitempty"`
}

// ImageGenType represents the type of image generation
type ImageGenType string

const (
	ImageGenTypePortrait      ImageGenType = "portrait"
	ImageGenTypeEvidenceBoard ImageGenType = "evidence_board"
	ImageGenTypeComparison    ImageGenType = "comparison"
	ImageGenTypeReportFigure  ImageGenType = "report_figure"
)

// IsValid checks if the image gen type is valid
func (t ImageGenType) IsValid() bool {
	switch t {
	case ImageGenTypePortrait, ImageGenTypeEvidenceBoard, ImageGenTypeComparison, ImageGenTypeReportFigure:
		return true
	}
	return false
}

// Validate checks if the ImageGenInput is valid
func (i *ImageGenInput) Validate() error {
	if i.CaseID == "" {
		return errors.New("case_id is required")
	}
	if !i.GenType.IsValid() {
		return errors.New("invalid gen_type")
	}
	if i.Resolution != "1k" && i.Resolution != "2k" && i.Resolution != "4k" {
		return errors.New("resolution must be 1k, 2k, or 4k")
	}
	if i.GenType == ImageGenTypePortrait && i.PortraitAttrs == nil {
		return errors.New("portrait_attributes required for portrait generation")
	}
	return nil
}

// GetModelForResolution returns the appropriate Nano Banana model for the resolution
// See: https://ai.google.dev/gemini-api/docs/image-generation
func (i *ImageGenInput) GetModelForResolution() string {
	if i.Resolution == "2k" || i.Resolution == "4k" {
		return "gemini-3-pro-image-preview" // Nano Banana Pro for high quality
	}
	return "gemini-2.5-flash-image" // Nano Banana for fast iteration
}

// ImageGenOutput represents output from image generation jobs
type ImageGenOutput struct {
	AssetKey       string  `json:"asset_key"`
	ThumbnailKey   string  `json:"thumbnail_key"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	ModelUsed      string  `json:"model_used"`
	GenerationTime int64   `json:"generation_time_ms"`
	CostUSD        float64 `json:"cost_usd"`
}

// ProfileInput represents input for profile extraction jobs
type ProfileInput struct {
	CaseID             string                  `json:"case_id"`
	Statements         []WitnessStatementInput `json:"statements"`
	ExistingAttributes *SuspectAttributes      `json:"existing_attributes,omitempty"`
	CommitID           string                  `json:"commit_id,omitempty"`
}

// Validate checks if the ProfileInput is valid
func (p *ProfileInput) Validate() error {
	if p.CaseID == "" {
		return errors.New("case_id is required")
	}
	if len(p.Statements) == 0 {
		return errors.New("at least one statement is required")
	}
	for i, stmt := range p.Statements {
		if err := stmt.Validate(); err != nil {
			return errors.New("statement " + string(rune('0'+i)) + ": " + err.Error())
		}
	}
	return nil
}

// ProfileOutput represents output from profile extraction jobs
type ProfileOutput struct {
	Attributes         *SuspectAttributes       `json:"attributes"`
	ExtractedFacts     []ExtractedFact          `json:"extracted_facts"`
	Conflicts          []AttributeConflict      `json:"conflicts,omitempty"`
	ConfidenceChanges  map[string]float64       `json:"confidence_changes,omitempty"`
	ImageGenTriggered  bool                     `json:"imagegen_triggered"`
	ImageGenJobID      string                   `json:"imagegen_job_id,omitempty"`
}

// ExtractedFact represents a fact extracted from a statement
type ExtractedFact struct {
	Attribute   string  `json:"attribute"`
	Value       string  `json:"value"`
	SourceIndex int     `json:"source_index"`
	Confidence  float64 `json:"confidence"`
}

// AttributeConflict represents a conflict between attribute values
type AttributeConflict struct {
	Attribute     string   `json:"attribute"`
	Values        []string `json:"values"`
	SourceIndices []int    `json:"source_indices"`
	Resolution    string   `json:"resolution,omitempty"`
}

// ============================================
// REPLAY JOB (HY-World-1.5)
// ============================================

// ReplayInput represents input for trajectory replay/animation jobs
type ReplayInput struct {
	CaseID       string      `json:"case_id"`
	TrajectoryID string      `json:"trajectory_id"`          // ID of trajectory to replay
	Trajectory   *Trajectory `json:"trajectory,omitempty"`   // Or full trajectory data
	SceneImage   string      `json:"scene_image,omitempty"`  // Base scene image key
	Perspective  string      `json:"perspective"`            // "first_person" or "third_person"
	FrameCount   int         `json:"frame_count,omitempty"`  // Default 125 (24 FPS * ~5s)
	Resolution   string      `json:"resolution,omitempty"`   // "480p" or "720p"

	// Additional fields for Modal HY-WorldPlay integration
	ReferenceImageKey      string `json:"reference_image_key,omitempty"`      // Storage key for reference image
	SceneDescription       string `json:"scene_description,omitempty"`        // Text description of the scene
	TrajectoryDescription  string `json:"trajectory_description,omitempty"`   // Text description of movement
	CameraPose             string `json:"camera_pose,omitempty"`              // Camera pose commands (e.g., "w-31,right-10")
}

// Validate checks if the ReplayInput is valid
func (r *ReplayInput) Validate() error {
	if r.CaseID == "" {
		return errors.New("case_id is required")
	}
	if r.TrajectoryID == "" && r.Trajectory == nil {
		return errors.New("trajectory_id or trajectory data is required")
	}
	if r.Perspective == "" {
		r.Perspective = "third_person"
	}
	if r.Perspective != "first_person" && r.Perspective != "third_person" {
		return errors.New("perspective must be 'first_person' or 'third_person'")
	}
	return nil
}

// SetDefaults sets default values for ReplayInput
func (r *ReplayInput) SetDefaults() {
	if r.FrameCount == 0 {
		r.FrameCount = 125 // ~5 seconds at 24 FPS
	}
	if r.Resolution == "" {
		r.Resolution = "480p"
	}
	if r.Perspective == "" {
		r.Perspective = "third_person"
	}
}

// ReplayOutput represents output from replay jobs
type ReplayOutput struct {
	VideoAssetKey  string `json:"video_asset_key"`
	ThumbnailKey   string `json:"thumbnail_key,omitempty"`
	FrameCount     int    `json:"frame_count"`
	FPS            int    `json:"fps"`
	DurationMs     int64  `json:"duration_ms"`
	Resolution     string `json:"resolution"`
	ModelUsed      string `json:"model_used"` // "hy-world-1.5"
	GenerationTime int64  `json:"generation_time_ms"`
}

// ============================================
// ASSET 3D JOB (Hunyuan3D-2)
// ============================================

// Asset3DInput represents input for 3D asset generation jobs
type Asset3DInput struct {
	CaseID      string `json:"case_id"`
	ImageKey    string `json:"image_key"`              // Storage key of evidence photo
	Description string `json:"description,omitempty"`  // Optional text description
	ItemType    string `json:"item_type,omitempty"`    // "weapon", "footprint", "tool", "other"
	WithTexture bool   `json:"with_texture"`           // Generate textured model
	OutputFormat string `json:"output_format,omitempty"` // "glb", "obj", "ply" (default: glb)
}

// Validate checks if the Asset3DInput is valid
func (a *Asset3DInput) Validate() error {
	if a.CaseID == "" {
		return errors.New("case_id is required")
	}
	if a.ImageKey == "" {
		return errors.New("image_key is required")
	}
	return nil
}

// SetDefaults sets default values for Asset3DInput
func (a *Asset3DInput) SetDefaults() {
	if a.OutputFormat == "" {
		a.OutputFormat = "glb"
	}
	if a.ItemType == "" {
		a.ItemType = "other"
	}
}

// Asset3DOutput represents output from 3D asset generation jobs
type Asset3DOutput struct {
	MeshAssetKey   string `json:"mesh_asset_key"`
	ThumbnailKey   string `json:"thumbnail_key,omitempty"`
	Format         string `json:"format"`       // "glb", "obj", "ply"
	HasTexture     bool   `json:"has_texture"`
	VertexCount    int    `json:"vertex_count,omitempty"`
	ModelUsed      string `json:"model_used"`   // "hunyuan3d-2"
	GenerationTime int64  `json:"generation_time_ms"`
}

// ============================================
// SCENE ANALYSIS JOB (Gemini Vision)
// ============================================

// SceneAnalysisInput represents input for scene analysis jobs
type SceneAnalysisInput struct {
	CaseID    string   `json:"case_id"`
	ImageKeys []string `json:"image_keys"`           // Storage keys of images to analyze
	Query     string   `json:"query,omitempty"`      // Optional specific question
	Mode      string   `json:"mode,omitempty"`       // "object_detection", "evidence_search", "full_analysis"
}

// Validate checks if the SceneAnalysisInput is valid
func (s *SceneAnalysisInput) Validate() error {
	if s.CaseID == "" {
		return errors.New("case_id is required")
	}
	if len(s.ImageKeys) == 0 {
		return errors.New("at least one image_key is required")
	}
	return nil
}

// SetDefaults sets default values for SceneAnalysisInput
func (s *SceneAnalysisInput) SetDefaults() {
	if s.Mode == "" {
		s.Mode = "full_analysis"
	}
}

// SceneAnalysisOutput represents output from scene analysis jobs
type SceneAnalysisOutput struct {
	DetectedObjects   []DetectedObject   `json:"detected_objects"`
	PotentialEvidence []string           `json:"potential_evidence"`
	SceneDescription  string             `json:"scene_description"`
	Anomalies         []string           `json:"anomalies,omitempty"`
	AnalysisTime      int64              `json:"analysis_time_ms"`
	ModelUsed         string             `json:"model_used"` // "gemini-2.5-flash-vision"
}

// DetectedObject represents an object detected in scene analysis
type DetectedObject struct {
	ID                 string   `json:"id"`
	Type               string   `json:"type"`
	Label              string   `json:"label"`
	PositionDescription string  `json:"position_description"`
	BoundingBox        *BBox    `json:"bounding_box,omitempty"`
	Confidence         float64  `json:"confidence"`
	IsSuspicious       bool     `json:"is_suspicious"`
	Notes              string   `json:"notes,omitempty"`
	SourceImageKey     string   `json:"source_image_key"`
}

// BBox represents a 2D bounding box for detected objects
type BBox struct {
	X      float64 `json:"x"`      // Top-left X (0-1 normalized)
	Y      float64 `json:"y"`      // Top-left Y (0-1 normalized)
	Width  float64 `json:"width"`  // Width (0-1 normalized)
	Height float64 `json:"height"` // Height (0-1 normalized)
}
