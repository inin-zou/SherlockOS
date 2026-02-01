package workers

import (
	"context"
	"encoding/json"
	"fmt"

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
func NewReasoningWorker(database *db.DB, q *queue.Queue, client clients.ReasoningClient) *ReasoningWorker {
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
		return fmt.Errorf("failed to parse input: %w", err)
	}

	// Validate input
	if input.Scenegraph == nil {
		return fmt.Errorf("scenegraph is required for reasoning")
	}

	// Set defaults
	if input.ThinkingBudget == 0 {
		input.ThinkingBudget = 8192 // Default thinking budget
	}
	if input.MaxTrajectories == 0 {
		input.MaxTrajectories = 3 // Default top-K
	}

	// Update progress: starting
	w.UpdateJobProgress(ctx, job.JobID.String(), 10)

	// Call reasoning service (Gemini 2.5 Flash)
	output, err := w.client.Reason(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID.String(), err)
		return fmt.Errorf("reasoning failed: %w", err)
	}

	// Update progress: processing complete
	w.UpdateJobProgress(ctx, job.JobID.String(), 90)

	// TODO: Create commit with reasoning_result type

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID.String(), 100)
	w.MarkJobDone(ctx, job.JobID.String(), output)

	return nil
}
