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

// Asset3DWorker handles 3D asset generation jobs using Hunyuan3D-2
type Asset3DWorker struct {
	*BaseWorker
	client clients.Asset3DClient
}

// NewAsset3DWorker creates a new 3D asset generation worker
func NewAsset3DWorker(database *db.DB, q queue.JobQueue, client clients.Asset3DClient) *Asset3DWorker {
	return &Asset3DWorker{
		BaseWorker: NewBaseWorker(database, q),
		client:     client,
	}
}

// Type returns the job type this worker handles
func (w *Asset3DWorker) Type() models.JobType {
	return models.JobTypeAsset3D
}

// Process handles a 3D asset generation job
func (w *Asset3DWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.Asset3DInput
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

	// Call 3D asset generation service (Hunyuan3D-2 via Replicate)
	output, err := w.client.Generate3DAsset(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return NewRetryableError(fmt.Errorf("3D asset generation failed: %w", err))
	}

	// Update progress: generation complete
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Create asset record
	caseID, _ := uuid.Parse(input.CaseID)
	assetID, err := w.createAssetRecord(ctx, caseID, input, output)
	if err != nil {
		fmt.Printf("Warning: failed to create asset record: %v\n", err)
	}

	// Build full output with asset ID
	fullOutput := map[string]interface{}{
		"asset_id":        assetID,
		"mesh_asset_key":  output.MeshAssetKey,
		"thumbnail_key":   output.ThumbnailKey,
		"format":          output.Format,
		"has_texture":     output.HasTexture,
		"vertex_count":    output.VertexCount,
		"model_used":      output.ModelUsed,
		"generation_time": output.GenerationTime,
	}

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, fullOutput)

	return nil
}

// createAssetRecord creates an asset record for the generated 3D model
func (w *Asset3DWorker) createAssetRecord(ctx context.Context, caseID uuid.UUID, input models.Asset3DInput, output *models.Asset3DOutput) (string, error) {
	if w.repo == nil {
		return "", nil
	}

	asset := &models.Asset{
		ID:         uuid.New(),
		CaseID:     caseID,
		Kind:       models.AssetKindEvidenceModel,
		StorageKey: output.MeshAssetKey,
		Metadata: map[string]interface{}{
			"format":          output.Format,
			"has_texture":     output.HasTexture,
			"vertex_count":    output.VertexCount,
			"model_used":      output.ModelUsed,
			"generation_time": output.GenerationTime,
			"source_image":    input.ImageKey,
			"item_type":       input.ItemType,
			"description":     input.Description,
		},
		CreatedAt: time.Now().UTC(),
	}

	if err := w.repo.CreateAsset(ctx, asset); err != nil {
		return "", err
	}

	return asset.ID.String(), nil
}
