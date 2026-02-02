package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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
func NewReconstructionWorker(database *db.DB, q queue.JobQueue, client clients.ReconstructionClient) *ReconstructionWorker {
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
		return NewFatalError(fmt.Errorf("failed to parse input: %w", err))
	}

	// Validate input
	if err := input.Validate(); err != nil {
		return NewFatalError(fmt.Errorf("invalid input: %w", err))
	}

	// Update progress: starting
	w.UpdateJobProgress(ctx, job.JobID, 10)

	// Call reconstruction service
	output, err := w.client.Reconstruct(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return NewRetryableError(fmt.Errorf("reconstruction failed: %w", err))
	}

	// Update progress: processing complete
	w.UpdateJobProgress(ctx, job.JobID, 60)

	// Get existing SceneGraph or create empty one
	existingSG := input.ExistingScenegraph
	if existingSG == nil {
		existingSG = models.NewEmptySceneGraph()
	}

	// Merge reconstruction output into SceneGraph
	newSG := w.mergeReconstructionOutput(existingSG, output)

	// Update progress: merging complete
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Create commit with reconstruction_update type
	caseID, _ := uuid.Parse(input.CaseID)
	if err := w.createReconstructionCommit(ctx, caseID, job.JobID, output, newSG); err != nil {
		// Log but don't fail - reconstruction succeeded
		fmt.Printf("Warning: failed to create commit: %v\n", err)
	}

	// Update scene_snapshot with new SceneGraph
	if err := w.updateSceneSnapshot(ctx, caseID, newSG); err != nil {
		fmt.Printf("Warning: failed to update scene snapshot: %v\n", err)
	}

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, output)

	return nil
}

// mergeReconstructionOutput merges reconstruction output into existing SceneGraph
func (w *ReconstructionWorker) mergeReconstructionOutput(existing *models.SceneGraph, output *models.ReconstructionOutput) *models.SceneGraph {
	// Create a copy
	result := &models.SceneGraph{
		Version:            existing.Version,
		Bounds:             existing.Bounds,
		Objects:            make([]models.SceneObject, 0, len(existing.Objects)),
		Evidence:           make([]models.EvidenceCard, len(existing.Evidence)),
		Constraints:        make([]models.Constraint, len(existing.Constraints)),
		UncertaintyRegions: output.UncertaintyRegions,
	}

	// Copy existing objects into map for lookup
	objectMap := make(map[string]models.SceneObject)
	for _, obj := range existing.Objects {
		objectMap[obj.ID] = obj
	}

	// Process proposals
	for _, proposal := range output.Objects {
		switch proposal.Action {
		case "create":
			if proposal.Object != nil {
				objectMap[proposal.Object.ID] = *proposal.Object
			}
		case "update":
			if proposal.Object != nil {
				objectMap[proposal.Object.ID] = *proposal.Object
			}
		case "remove":
			delete(objectMap, proposal.ID)
		}
	}

	// Convert map back to slice
	for _, obj := range objectMap {
		result.Objects = append(result.Objects, obj)
	}

	// Copy evidence and constraints
	copy(result.Evidence, existing.Evidence)
	copy(result.Constraints, existing.Constraints)

	return result
}

// createReconstructionCommit creates a commit for the reconstruction update
func (w *ReconstructionWorker) createReconstructionCommit(ctx context.Context, caseID, jobID uuid.UUID, output *models.ReconstructionOutput, newSG *models.SceneGraph) error {
	if w.repo == nil {
		return nil
	}

	// Build changes summary
	var added, updated, removed []string
	for _, proposal := range output.Objects {
		switch proposal.Action {
		case "create":
			if proposal.Object != nil {
				added = append(added, proposal.Object.ID)
			}
		case "update":
			updated = append(updated, proposal.ID)
		case "remove":
			removed = append(removed, proposal.ID)
		}
	}

	payload := map[string]interface{}{
		"job_id":     jobID.String(),
		"scenegraph": newSG,
		"changes": map[string]interface{}{
			"objects_added":   added,
			"objects_updated": updated,
			"objects_removed": removed,
		},
		"processing_stats": output.ProcessingStats,
	}

	summary := fmt.Sprintf("Scene reconstruction: %d objects detected", output.ProcessingStats.DetectedObjects)

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

// updateSceneSnapshot updates the scene snapshot with new SceneGraph
func (w *ReconstructionWorker) updateSceneSnapshot(ctx context.Context, caseID uuid.UUID, sg *models.SceneGraph) error {
	if w.repo == nil {
		return nil
	}

	// Get latest commit ID
	var commitID uuid.UUID
	latestCommit, _ := w.repo.GetLatestCommit(ctx, caseID)
	if latestCommit != nil {
		commitID = latestCommit.ID
	}

	snapshot := &models.SceneSnapshot{
		CaseID:     caseID,
		CommitID:   commitID,
		Scenegraph: sg,
		UpdatedAt:  time.Now().UTC(),
	}

	return w.repo.UpsertSceneSnapshot(ctx, snapshot)
}
