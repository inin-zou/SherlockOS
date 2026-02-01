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
	CaseID            string       `json:"case_id"`
	ScanAssetKeys     []string     `json:"scan_asset_keys"`
	CameraPoses       []CameraPose `json:"camera_poses,omitempty"`
	DepthMaps         []string     `json:"depth_maps,omitempty"`
	ExistingScenegraph *SceneGraph `json:"existing_scenegraph,omitempty"`
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
