package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

// ReplayWorker handles trajectory replay video generation jobs using HY-World-1.5
type ReplayWorker struct {
	*BaseWorker
	client clients.ReplayClient
}

// NewReplayWorker creates a new replay video generation worker
func NewReplayWorker(database *db.DB, q queue.JobQueue, client clients.ReplayClient) *ReplayWorker {
	return &ReplayWorker{
		BaseWorker: NewBaseWorker(database, q),
		client:     client,
	}
}

// Type returns the job type this worker handles
func (w *ReplayWorker) Type() models.JobType {
	return models.JobTypeReplay
}

// Process handles a trajectory replay video generation job
func (w *ReplayWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.ReplayInput
	if err := json.Unmarshal(job.Input, &input); err != nil {
		return NewFatalError(fmt.Errorf("failed to parse input: %w", err))
	}

	// Validate and set defaults
	if err := input.Validate(); err != nil {
		return NewFatalError(fmt.Errorf("invalid input: %w", err))
	}
	input.SetDefaults()

	// Update progress: starting
	w.UpdateJobProgress(ctx, job.JobID, 10)

	// Call replay generation service (HY-World-1.5 via Modal)
	output, err := w.client.GenerateReplay(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		errMsg := err.Error()
		// Don't retry timeout/unavailable errors - they won't succeed on retry
		if strings.Contains(errMsg, "204 No Content") ||
		   strings.Contains(errMsg, "timed out") ||
		   strings.Contains(errMsg, "context deadline exceeded") ||
		   strings.Contains(errMsg, "function execution timed out") ||
		   strings.Contains(errMsg, "status 500") {
			return NewFatalError(fmt.Errorf("replay generation failed: %w", err))
		}
		return NewRetryableError(fmt.Errorf("replay generation failed: %w", err))
	}

	// Update progress: generation complete
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Create asset record for the video
	caseID, _ := uuid.Parse(input.CaseID)
	assetID, err := w.createAssetRecord(ctx, caseID, input, output)
	if err != nil {
		fmt.Printf("Warning: failed to create asset record: %v\n", err)
	}

	// Create commit for the replay generation
	commitID, err := w.createCommit(ctx, caseID, job.JobID, input, output)
	if err != nil {
		fmt.Printf("Warning: failed to create commit: %v\n", err)
	}

	// Build full output
	fullOutput := map[string]interface{}{
		"asset_id":        assetID,
		"commit_id":       commitID,
		"video_asset_key": output.VideoAssetKey,
		"thumbnail_key":   output.ThumbnailKey,
		"frame_count":     output.FrameCount,
		"fps":             output.FPS,
		"duration_ms":     output.DurationMs,
		"resolution":      output.Resolution,
		"model_used":      output.ModelUsed,
		"generation_time": output.GenerationTime,
	}

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, fullOutput)

	return nil
}

// createAssetRecord creates an asset record for the generated replay video
func (w *ReplayWorker) createAssetRecord(ctx context.Context, caseID uuid.UUID, input models.ReplayInput, output *models.ReplayOutput) (string, error) {
	if w.repo == nil {
		return "", nil
	}

	asset := &models.Asset{
		ID:         uuid.New(),
		CaseID:     caseID,
		Kind:       models.AssetKindReplayVideo,
		StorageKey: output.VideoAssetKey,
		Metadata: map[string]interface{}{
			"trajectory_id":   input.TrajectoryID,
			"frame_count":     output.FrameCount,
			"fps":             output.FPS,
			"duration_ms":     output.DurationMs,
			"resolution":      output.Resolution,
			"model_used":      output.ModelUsed,
			"generation_time": output.GenerationTime,
			"perspective":     input.Perspective,
		},
		CreatedAt: time.Now().UTC(),
	}

	if err := w.repo.CreateAsset(ctx, asset); err != nil {
		return "", err
	}

	return asset.ID.String(), nil
}

// createCommit creates a commit record for the replay generation
func (w *ReplayWorker) createCommit(ctx context.Context, caseID uuid.UUID, jobID uuid.UUID, input models.ReplayInput, output *models.ReplayOutput) (string, error) {
	if w.repo == nil {
		return "", nil
	}

	payload := map[string]interface{}{
		"job_id":          jobID.String(),
		"trajectory_id":   input.TrajectoryID,
		"video_asset_key": output.VideoAssetKey,
		"frame_count":     output.FrameCount,
		"duration_ms":     output.DurationMs,
		"resolution":      output.Resolution,
	}

	commit, err := models.NewCommit(
		caseID,
		models.CommitTypeReplayGenerated,
		fmt.Sprintf("Generated replay video for trajectory %s", input.TrajectoryID),
		payload,
	)
	if err != nil {
		return "", err
	}

	if err := w.repo.CreateCommit(ctx, commit); err != nil {
		return "", err
	}

	return commit.ID.String(), nil
}
