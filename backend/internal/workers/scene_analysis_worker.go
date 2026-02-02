package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

// SceneAnalysisWorker handles scene analysis jobs using Gemini 3 Pro Vision
type SceneAnalysisWorker struct {
	*BaseWorker
	client clients.SceneAnalysisClient
}

// NewSceneAnalysisWorker creates a new scene analysis worker
func NewSceneAnalysisWorker(database *db.DB, q queue.JobQueue, client clients.SceneAnalysisClient) *SceneAnalysisWorker {
	return &SceneAnalysisWorker{
		BaseWorker: NewBaseWorker(database, q),
		client:     client,
	}
}

// Type returns the job type this worker handles
func (w *SceneAnalysisWorker) Type() models.JobType {
	return models.JobTypeSceneAnalysis
}

// Process handles a scene analysis job
func (w *SceneAnalysisWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.SceneAnalysisInput
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

	// Call scene analysis service (Gemini 3 Pro Vision)
	output, err := w.client.AnalyzeScene(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return NewRetryableError(fmt.Errorf("scene analysis failed: %w", err))
	}

	// Update progress: analysis complete
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Create commit with scene analysis results
	caseID, _ := uuid.Parse(input.CaseID)
	if err := w.createSceneAnalysisCommit(ctx, caseID, job.JobID, output); err != nil {
		fmt.Printf("Warning: failed to create scene analysis commit: %v\n", err)
	}

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, output)

	return nil
}

// createSceneAnalysisCommit creates a commit for scene analysis results
func (w *SceneAnalysisWorker) createSceneAnalysisCommit(ctx context.Context, caseID, jobID uuid.UUID, output *models.SceneAnalysisOutput) error {
	if w.repo == nil {
		return nil
	}

	payload := map[string]interface{}{
		"job_id":             jobID.String(),
		"detected_objects":   output.DetectedObjects,
		"potential_evidence": output.PotentialEvidence,
		"scene_description":  output.SceneDescription,
		"anomalies":          output.Anomalies,
		"model_used":         output.ModelUsed,
		"analysis_time_ms":   output.AnalysisTime,
	}

	// Count suspicious objects
	suspiciousCount := 0
	for _, obj := range output.DetectedObjects {
		if obj.IsSuspicious {
			suspiciousCount++
		}
	}

	summary := fmt.Sprintf("Scene analysis: detected %d objects", len(output.DetectedObjects))
	if suspiciousCount > 0 {
		summary += fmt.Sprintf(" (%d suspicious)", suspiciousCount)
	}
	if len(output.PotentialEvidence) > 0 {
		summary += fmt.Sprintf(", %d potential evidence items", len(output.PotentialEvidence))
	}

	// Use reconstruction_update commit type since this updates the scene understanding
	commit, err := models.NewCommit(caseID, models.CommitTypeReconstructionUpdate, summary, payload)
	if err != nil {
		return err
	}

	// Get latest commit as parent
	latestCommit, _ := w.repo.GetLatestCommit(ctx, caseID)
	if latestCommit != nil {
		commit.SetParent(latestCommit.ID)
	}

	return w.repo.CreateCommit(ctx, commit)
}
