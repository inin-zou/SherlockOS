package clients

import (
	"context"

	"github.com/sherlockos/backend/internal/models"
)

// ReconstructionClient defines the interface for scene reconstruction services
type ReconstructionClient interface {
	// Reconstruct processes scan images and returns scene updates
	Reconstruct(ctx context.Context, input models.ReconstructionInput) (*models.ReconstructionOutput, error)
}

// ImageGenClient defines the interface for image generation services
type ImageGenClient interface {
	// Generate creates an image based on the input parameters
	Generate(ctx context.Context, input ImageGenInput) (*ImageGenOutput, error)
}

// ImageGenInput represents input for image generation
type ImageGenInput struct {
	CaseID            string                    `json:"case_id"`
	GenType           string                    `json:"gen_type"` // "portrait", "evidence_board", "comparison", "report_figure"
	PortraitAttrs     *models.SuspectAttributes `json:"portrait_attributes,omitempty"`
	ReferenceImageKey string                    `json:"reference_image_key,omitempty"`
	ObjectIDs         []string                  `json:"object_ids,omitempty"`
	Layout            string                    `json:"layout,omitempty"` // "grid", "timeline", "comparison"
	Resolution        string                    `json:"resolution"`       // "1k", "2k", "4k"
	StylePrompt       string                    `json:"style_prompt,omitempty"`
}

// ImageGenOutput represents output from image generation
type ImageGenOutput struct {
	AssetKey       string `json:"asset_key"`
	ThumbnailKey   string `json:"thumbnail_key"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	ModelUsed      string `json:"model_used"` // "nano-banana", "nano-banana-pro"
	GenerationTime int64  `json:"generation_time_ms"`
	CostUSD        float64 `json:"cost_usd"`
}

// ReasoningClient defines the interface for trajectory reasoning services
type ReasoningClient interface {
	// Reason generates trajectory hypotheses based on scene data
	Reason(ctx context.Context, input models.ReasoningInput) (*models.ReasoningOutput, error)
}

// ProfileClient defines the interface for suspect profile extraction
type ProfileClient interface {
	// ExtractProfile parses witness statements and updates attributes
	ExtractProfile(ctx context.Context, statements []models.WitnessStatementInput, existing *models.SuspectAttributes) (*models.SuspectAttributes, error)
}

// StorageClient defines the interface for file storage operations
type StorageClient interface {
	// GenerateUploadURL creates a presigned URL for uploading
	GenerateUploadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error)

	// GenerateDownloadURL creates a presigned URL for downloading
	GenerateDownloadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, bucket, key string) error
}
