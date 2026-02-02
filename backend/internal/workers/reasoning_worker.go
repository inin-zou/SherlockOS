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

// ReasoningWorker handles trajectory reasoning jobs
type ReasoningWorker struct {
	*BaseWorker
	client clients.ReasoningClient
}

// NewReasoningWorker creates a new reasoning worker
func NewReasoningWorker(database *db.DB, q queue.JobQueue, client clients.ReasoningClient) *ReasoningWorker {
	return &ReasoningWorker{
		BaseWorker: NewBaseWorker(database, q),
		client:     client,
	}
}

// Type returns the job type this worker handles
func (w *ReasoningWorker) Type() models.JobType {
	return models.JobTypeReasoning
}

// Process handles a reasoning job
func (w *ReasoningWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.ReasoningInput
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

	// Call reasoning service (Gemini 2.5 Flash with Thinking)
	output, err := w.client.Reason(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return NewRetryableError(fmt.Errorf("reasoning failed: %w", err))
	}

	// Update progress: processing complete
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Create commit with reasoning_result type
	caseID, _ := uuid.Parse(input.CaseID)
	if err := w.createReasoningCommit(ctx, caseID, job.JobID, input, output); err != nil {
		fmt.Printf("Warning: failed to create reasoning commit: %v\n", err)
	}

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, output)

	return nil
}

// createReasoningCommit creates a commit for reasoning results
func (w *ReasoningWorker) createReasoningCommit(ctx context.Context, caseID, jobID uuid.UUID, input models.ReasoningInput, output *models.ReasoningOutput) error {
	if w.repo == nil {
		return nil
	}

	payload := map[string]interface{}{
		"job_id":              jobID.String(),
		"trajectories":        output.Trajectories,
		"uncertainty_areas":   output.UncertaintyAreas,
		"suggestions":         output.NextStepSuggestions,
		"thinking_summary":    output.ThinkingSummary,
		"model_stats":         output.ModelStats,
		"thinking_budget":     input.ThinkingBudget,
		"max_trajectories":    input.MaxTrajectories,
	}

	// Add branch ID if present
	if input.BranchID != "" {
		payload["branch_id"] = input.BranchID
	}

	summary := fmt.Sprintf("Generated %d trajectory hypotheses", len(output.Trajectories))
	if len(output.Trajectories) > 0 {
		summary += fmt.Sprintf(" (top confidence: %.1f%%)", output.Trajectories[0].OverallConfidence*100)
	}

	commit, err := models.NewCommit(caseID, models.CommitTypeReasoningResult, summary, payload)
	if err != nil {
		return err
	}

	// Get latest commit as parent
	latestCommit, _ := w.repo.GetLatestCommit(ctx, caseID)
	if latestCommit != nil {
		commit.SetParent(latestCommit.ID)
	}

	// Set branch if specified
	if input.BranchID != "" {
		branchUUID, err := uuid.Parse(input.BranchID)
		if err == nil {
			commit.SetBranch(branchUUID)
		}
	}

	return w.repo.CreateCommit(ctx, commit)
}
