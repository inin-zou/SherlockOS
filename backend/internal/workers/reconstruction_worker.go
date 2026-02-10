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

	// Check if preprocessing is requested but POV images not yet generated
	if input.EnablePreprocess && len(input.GeneratedPOVKeys) == 0 {
		fmt.Printf("Reconstruction job %s: preprocessing enabled, generating POV images first\n", job.JobID)

		// Update progress: preprocessing
		w.UpdateJobProgress(ctx, job.JobID, 5)

		// Generate POV images
		povKeys, err := w.generatePOVImages(ctx, job.JobID, &input)
		if err != nil {
			// Log warning but continue without POV images
			fmt.Printf("Warning: POV generation failed: %v (continuing with raw images only)\n", err)
		} else {
			input.GeneratedPOVKeys = povKeys
			fmt.Printf("POV generation complete: %d images generated\n", len(povKeys))
		}
	}

	// Log input details
	fmt.Printf("Reconstruction job %s: %d raw images", job.JobID, len(input.ScanAssetKeys))
	if len(input.GeneratedPOVKeys) > 0 {
		fmt.Printf(", %d generated POV images", len(input.GeneratedPOVKeys))
	}
	fmt.Println()

	// Update progress: starting reconstruction
	w.UpdateJobProgress(ctx, job.JobID, 20)

	// Call reconstruction service (will combine raw + POV images internally)
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
	if err := w.createReconstructionCommit(ctx, caseID, job.JobID, &input, output, newSG); err != nil {
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
		PointCloud:         output.PointCloud, // Pass through point cloud from reconstruction
		GaussianAssetKey:   output.GaussianAssetKey,
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

	// Compute bounds from all objects
	result.Bounds = w.computeBoundsFromObjects(result.Objects)

	return result
}

// computeBoundsFromObjects calculates the scene bounds from all object bounding boxes
func (w *ReconstructionWorker) computeBoundsFromObjects(objects []models.SceneObject) models.BoundingBox {
	if len(objects) == 0 {
		// Default room bounds if no objects
		return models.BoundingBox{
			Min: [3]float64{-7, 0, -6},
			Max: [3]float64{7, 4, 6},
		}
	}

	// Initialize with extreme values
	minX, minY, minZ := 1e9, 0.0, 1e9
	maxX, maxY, maxZ := -1e9, 4.0, -1e9 // minY starts at floor, maxY at typical ceiling

	for _, obj := range objects {
		// Use object position
		pos := obj.Pose.Position

		// Consider object bounding box if available
		if obj.BBox.Min != [3]float64{0, 0, 0} || obj.BBox.Max != [3]float64{0, 0, 0} {
			if obj.BBox.Min[0]+pos[0] < minX {
				minX = obj.BBox.Min[0] + pos[0]
			}
			if obj.BBox.Min[2]+pos[2] < minZ {
				minZ = obj.BBox.Min[2] + pos[2]
			}
			if obj.BBox.Max[0]+pos[0] > maxX {
				maxX = obj.BBox.Max[0] + pos[0]
			}
			if obj.BBox.Max[1]+pos[1] > maxY {
				maxY = obj.BBox.Max[1] + pos[1]
			}
			if obj.BBox.Max[2]+pos[2] > maxZ {
				maxZ = obj.BBox.Max[2] + pos[2]
			}
		} else {
			// Use position directly with some margin
			if pos[0]-1 < minX {
				minX = pos[0] - 1
			}
			if pos[2]-1 < minZ {
				minZ = pos[2] - 1
			}
			if pos[0]+1 > maxX {
				maxX = pos[0] + 1
			}
			if pos[2]+1 > maxZ {
				maxZ = pos[2] + 1
			}
		}
	}

	// Add margins for room walls (2m around objects)
	margin := 2.0
	minX -= margin
	minZ -= margin
	maxX += margin
	maxZ += margin

	// Ensure minimum room size
	if maxX-minX < 8 {
		center := (maxX + minX) / 2
		minX = center - 4
		maxX = center + 4
	}
	if maxZ-minZ < 8 {
		center := (maxZ + minZ) / 2
		minZ = center - 4
		maxZ = center + 4
	}
	if maxY < 3 {
		maxY = 3.5
	}

	return models.BoundingBox{
		Min: [3]float64{minX, minY, minZ},
		Max: [3]float64{maxX, maxY, maxZ},
	}
}

// createReconstructionCommit creates a commit for the reconstruction update
func (w *ReconstructionWorker) createReconstructionCommit(ctx context.Context, caseID, jobID uuid.UUID, input *models.ReconstructionInput, output *models.ReconstructionOutput, newSG *models.SceneGraph) error {
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

	// Track input sources
	inputSources := map[string]interface{}{
		"raw_image_count": len(input.ScanAssetKeys),
		"raw_image_keys":  input.ScanAssetKeys,
	}
	if len(input.GeneratedPOVKeys) > 0 {
		inputSources["pov_image_count"] = len(input.GeneratedPOVKeys)
		inputSources["pov_image_keys"] = input.GeneratedPOVKeys
		inputSources["hybrid_mode"] = true
	}

	payload := map[string]interface{}{
		"job_id":     jobID.String(),
		"scenegraph": newSG,
		"changes": map[string]interface{}{
			"objects_added":   added,
			"objects_updated": updated,
			"objects_removed": removed,
		},
		"input_sources":    inputSources,
		"processing_stats": output.ProcessingStats,
	}

	// Build summary with hybrid mode indicator
	summary := fmt.Sprintf("Scene reconstruction: %d objects detected", output.ProcessingStats.DetectedObjects)
	if len(input.GeneratedPOVKeys) > 0 {
		summary += fmt.Sprintf(" (hybrid: %d raw + %d POV images)", len(input.ScanAssetKeys), len(input.GeneratedPOVKeys))
	}

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

// generatePOVImages creates a POV generation job and waits for it to complete
func (w *ReconstructionWorker) generatePOVImages(ctx context.Context, parentJobID uuid.UUID, input *models.ReconstructionInput) ([]string, error) {
	if w.repo == nil || w.queue == nil {
		return nil, fmt.Errorf("repository or queue not available")
	}

	// Validate scene description is provided
	if input.SceneDescription == "" {
		return nil, fmt.Errorf("scene_description is required for POV generation")
	}

	// Build POV generation input
	caseID, err := uuid.Parse(input.CaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid case_id: %w", err)
	}

	// Default view angles for reconstruction
	viewAngles := []string{"front", "left", "right", "back"}

	povInput := models.ImageGenInput{
		CaseID:           input.CaseID,
		GenType:          models.ImageGenTypeScenePOV,
		SceneDescription: input.SceneDescription,
		ViewAngles:       viewAngles,
		RoomType:         input.RoomType,
		Resolution:       "1k", // Use 1k for speed during preprocessing
	}

	// Create the POV generation job
	povJob, err := models.NewJob(caseID, models.JobTypeImageGen, povInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create POV job: %w", err)
	}

	// Save job to database
	if err := w.repo.CreateJob(ctx, povJob); err != nil {
		return nil, fmt.Errorf("failed to save POV job: %w", err)
	}

	fmt.Printf("Created POV generation sub-job %s for reconstruction job %s\n", povJob.ID, parentJobID)

	// Enqueue the job
	if err := w.queue.Enqueue(ctx, povJob); err != nil {
		return nil, fmt.Errorf("failed to enqueue POV job: %w", err)
	}

	// Poll for job completion
	return w.waitForPOVJobCompletion(ctx, povJob.ID)
}

// waitForPOVJobCompletion polls the job status until it completes or fails
func (w *ReconstructionWorker) waitForPOVJobCompletion(ctx context.Context, jobID uuid.UUID) ([]string, error) {
	pollInterval := 2 * time.Second
	maxWait := 5 * time.Minute
	startTime := time.Now()

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check timeout
		if time.Since(startTime) > maxWait {
			return nil, fmt.Errorf("POV generation timed out after %v", maxWait)
		}

		// Get job status
		job, err := w.repo.GetJob(ctx, jobID)
		if err != nil {
			return nil, fmt.Errorf("failed to get POV job status: %w", err)
		}

		switch job.Status {
		case models.JobStatusDone:
			// Extract POV keys from output
			return w.extractPOVKeysFromOutput(job.Output)

		case models.JobStatusFailed:
			return nil, fmt.Errorf("POV generation failed: %s", job.Error)

		case models.JobStatusCanceled:
			return nil, fmt.Errorf("POV generation was canceled")

		case models.JobStatusQueued, models.JobStatusRunning:
			// Still processing, wait and retry
			fmt.Printf("POV job %s status: %s, progress: %d%%\n", jobID, job.Status, job.Progress)
			time.Sleep(pollInterval)
		}
	}
}

// extractPOVKeysFromOutput extracts the generated POV image keys from job output
func (w *ReconstructionWorker) extractPOVKeysFromOutput(output json.RawMessage) ([]string, error) {
	if output == nil {
		return nil, fmt.Errorf("POV job output is nil")
	}

	var result struct {
		GeneratedImages []struct {
			AssetKey string `json:"asset_key"`
		} `json:"generated_images"`
		AssetKey string `json:"asset_key"` // Fallback for single image
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse POV job output: %w", err)
	}

	var keys []string

	// Prefer generated_images array (multi-image output)
	if len(result.GeneratedImages) > 0 {
		for _, img := range result.GeneratedImages {
			if img.AssetKey != "" {
				keys = append(keys, img.AssetKey)
			}
		}
	} else if result.AssetKey != "" {
		// Fallback to single asset_key
		keys = append(keys, result.AssetKey)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no POV images found in job output")
	}

	return keys, nil
}
