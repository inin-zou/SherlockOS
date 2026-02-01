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

// ReconstructionWorker handles scene reconstruction jobs
type ReconstructionWorker struct {
	*BaseWorker
	client clients.ReconstructionClient
}

// NewReconstructionWorker creates a new reconstruction worker
func NewReconstructionWorker(database *db.DB, q *queue.Queue, client clients.ReconstructionClient) *ReconstructionWorker {
	return &ReconstructionWorker{
		BaseWorker: NewBaseWorker(database, q),
		client:     client,
	}
}

// Type returns the job type this worker handles
func (w *ReconstructionWorker) Type() models.JobType {
	return models.JobTypeReconstruction
}

// Process handles a reconstruction job
func (w *ReconstructionWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.ReconstructionInput
	if err := json.Unmarshal(job.Input, &input); err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	// Update progress: starting
	w.UpdateJobProgress(ctx, job.JobID, 10)

	// Call reconstruction service
	output, err := w.client.Reconstruct(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return fmt.Errorf("reconstruction failed: %w", err)
	}

	// Update progress: processing complete
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// TODO: Create commit with reconstruction_update type
	// TODO: Update scene_snapshot with new SceneGraph

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, output)

	return nil
}
