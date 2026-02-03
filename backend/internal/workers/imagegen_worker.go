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

// ImageGenWorker handles image generation jobs
type ImageGenWorker struct {
	*BaseWorker
	client clients.ImageGenClient
}

// NewImageGenWorker creates a new image generation worker
func NewImageGenWorker(database *db.DB, q queue.JobQueue, client clients.ImageGenClient) *ImageGenWorker {
	return &ImageGenWorker{
		BaseWorker: NewBaseWorker(database, q),
		client:     client,
	}
}

// Type returns the job type this worker handles
func (w *ImageGenWorker) Type() models.JobType {
	return models.JobTypeImageGen
}

// Process handles an image generation job
func (w *ImageGenWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.ImageGenInput
	if err := json.Unmarshal(job.Input, &input); err != nil {
		return NewFatalError(fmt.Errorf("failed to parse input: %w", err))
	}

	// Validate input
	if err := input.Validate(); err != nil {
		return NewFatalError(fmt.Errorf("invalid input: %w", err))
	}

	// Update progress: starting
	w.UpdateJobProgress(ctx, job.JobID, 10)

	// Generate image
	output, err := w.client.Generate(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return NewRetryableError(fmt.Errorf("image generation failed: %w", err))
	}

	// Update progress: generation complete
	w.UpdateJobProgress(ctx, job.JobID, 70)

	// Create asset record
	caseID, _ := uuid.Parse(input.CaseID)
	assetID, err := w.createAssetRecord(ctx, caseID, input, output)
	if err != nil {
		fmt.Printf("Warning: failed to create asset record: %v\n", err)
	}

	// Update portrait asset key if this is a portrait
	if input.GenType == models.ImageGenTypePortrait {
		if err := w.updatePortraitAssetKey(ctx, caseID, output.AssetKey); err != nil {
			fmt.Printf("Warning: failed to update portrait asset key: %v\n", err)
		}
	}

	// Update progress: complete
	w.UpdateJobProgress(ctx, job.JobID, 100)

	// Add asset ID to output for reference
	outputWithAsset := map[string]interface{}{
		"asset_id":        assetID,
		"asset_key":       output.AssetKey,
		"thumbnail_key":   output.ThumbnailKey,
		"width":           output.Width,
		"height":          output.Height,
		"model_used":      output.ModelUsed,
		"generation_time": output.GenerationTime,
		"cost_usd":        output.CostUSD,
	}

	w.MarkJobDone(ctx, job.JobID, outputWithAsset)

	return nil
}

// createAssetRecord creates an asset record in the database
func (w *ImageGenWorker) createAssetRecord(ctx context.Context, caseID uuid.UUID, input models.ImageGenInput, output *models.ImageGenOutput) (string, error) {
	if w.repo == nil {
		return "", nil
	}

	// Handle POV generation (multiple assets)
	if input.GenType == models.ImageGenTypeScenePOV && len(output.GeneratedImages) > 0 {
		return w.createPOVAssetRecords(ctx, caseID, input, output)
	}

	// Determine asset kind based on gen type
	var kind models.AssetKind
	switch input.GenType {
	case models.ImageGenTypePortrait:
		kind = models.AssetKindPortrait
	case models.ImageGenTypeAssetClean:
		kind = models.AssetKindGeneratedImage
	default:
		kind = models.AssetKindGeneratedImage
	}

	asset := &models.Asset{
		ID:         uuid.New(),
		CaseID:     caseID,
		Kind:       kind,
		StorageKey: output.AssetKey,
		Metadata: map[string]interface{}{
			"width":           output.Width,
			"height":          output.Height,
			"model_used":      output.ModelUsed,
			"generation_time": output.GenerationTime,
			"cost_usd":        output.CostUSD,
			"gen_type":        input.GenType,
			"resolution":      input.Resolution,
			"thumbnail_key":   output.ThumbnailKey,
		},
		CreatedAt: time.Now().UTC(),
	}

	if err := w.repo.CreateAsset(ctx, asset); err != nil {
		return "", err
	}

	return asset.ID.String(), nil
}

// createPOVAssetRecords creates asset records for all generated POV images
func (w *ImageGenWorker) createPOVAssetRecords(ctx context.Context, caseID uuid.UUID, input models.ImageGenInput, output *models.ImageGenOutput) (string, error) {
	var assetIDs []string

	for _, genImg := range output.GeneratedImages {
		asset := &models.Asset{
			ID:         uuid.New(),
			CaseID:     caseID,
			Kind:       models.AssetKindGeneratedImage,
			StorageKey: genImg.AssetKey,
			Metadata: map[string]interface{}{
				"width":         genImg.Width,
				"height":        genImg.Height,
				"gen_type":      input.GenType,
				"view_angle":    genImg.ViewAngle,
				"thumbnail_key": genImg.ThumbnailKey,
				"purpose":       "reconstruction_pov",
			},
			CreatedAt: time.Now().UTC(),
		}

		if err := w.repo.CreateAsset(ctx, asset); err != nil {
			fmt.Printf("Warning: failed to create asset for %s view: %v\n", genImg.ViewAngle, err)
			continue
		}

		assetIDs = append(assetIDs, asset.ID.String())
	}

	if len(assetIDs) == 0 {
		return "", fmt.Errorf("failed to create any asset records")
	}

	// Return comma-separated list of asset IDs
	return strings.Join(assetIDs, ","), nil
}

// updatePortraitAssetKey updates the portrait asset key in suspect profile
func (w *ImageGenWorker) updatePortraitAssetKey(ctx context.Context, caseID uuid.UUID, assetKey string) error {
	if w.repo == nil {
		return nil
	}

	// Get existing profile
	profile, err := w.repo.GetSuspectProfile(ctx, caseID)
	if err != nil {
		return err
	}

	if profile == nil {
		// Create new profile with just the portrait
		profile = &models.SuspectProfile{
			CaseID:           caseID,
			CommitID:         uuid.Nil, // No commit associated
			Attributes:       models.NewEmptySuspectAttributes(),
			PortraitAssetKey: assetKey,
			UpdatedAt:        time.Now().UTC(),
		}
	} else {
		profile.PortraitAssetKey = assetKey
		profile.UpdatedAt = time.Now().UTC()
	}

	return w.repo.UpsertSuspectProfile(ctx, profile)
}
