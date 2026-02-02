package clients

import (
	"context"

	"github.com/sherlockos/backend/internal/models"
)

// ============================================
// SCENE RECONSTRUCTION (HunyuanWorld-Mirror)
// ============================================

// ReconstructionClient defines the interface for scene reconstruction services
// Implementation: HunyuanWorld-Mirror (multi-view photos → 3D scene)
type ReconstructionClient interface {
	// Reconstruct processes scan images and returns scene updates
	Reconstruct(ctx context.Context, input models.ReconstructionInput) (*models.ReconstructionOutput, error)
}

// ============================================
// IMAGE GENERATION (Nano Banana)
// ============================================

// ImageGenClient defines the interface for image generation services
// Implementation: Nano Banana (gemini-2.5-flash-image) / Nano Banana Pro (gemini-3-pro-image-preview)
type ImageGenClient interface {
	// Generate creates an image based on the input parameters
	Generate(ctx context.Context, input models.ImageGenInput) (*models.ImageGenOutput, error)
}

// ============================================
// TRAJECTORY REASONING (Gemini 2.5 Flash)
// ============================================

// ReasoningClient defines the interface for trajectory reasoning services
// Implementation: Gemini 2.5 Flash with Thinking mode
type ReasoningClient interface {
	// Reason generates trajectory hypotheses based on scene data
	Reason(ctx context.Context, input models.ReasoningInput) (*models.ReasoningOutput, error)
}

// ============================================
// PROFILE EXTRACTION (Gemini 2.5 Flash)
// ============================================

// ProfileClient defines the interface for suspect profile extraction
// Implementation: Gemini 2.5 Flash
type ProfileClient interface {
	// ExtractProfile parses witness statements and updates attributes
	ExtractProfile(ctx context.Context, statements []models.WitnessStatementInput, existing *models.SuspectAttributes) (*models.SuspectAttributes, error)
}

// ============================================
// TRAJECTORY REPLAY (HY-World-1.5)
// ============================================

// ReplayClient defines the interface for trajectory replay/animation services
// Implementation: HY-World-1.5 (WorldPlay) - trajectory → video animation
type ReplayClient interface {
	// GenerateReplay creates a video animation of a trajectory through the scene
	GenerateReplay(ctx context.Context, input models.ReplayInput) (*models.ReplayOutput, error)
}

// ============================================
// 3D ASSET GENERATION (Hunyuan3D-2)
// ============================================

// Asset3DClient defines the interface for 3D asset generation services
// Implementation: Hunyuan3D-2 (single image → 3D mesh with texture)
type Asset3DClient interface {
	// Generate3DAsset creates a 3D model from an evidence photo
	Generate3DAsset(ctx context.Context, input models.Asset3DInput) (*models.Asset3DOutput, error)
}

// ============================================
// SCENE ANALYSIS (Gemini Vision)
// ============================================

// SceneAnalysisClient defines the interface for scene analysis/understanding
// Implementation: Gemini 2.0 Flash (gemini-2.0-flash-001) - fast, accurate vision model
type SceneAnalysisClient interface {
	// AnalyzeScene processes images and returns detected objects/evidence
	AnalyzeScene(ctx context.Context, input models.SceneAnalysisInput) (*models.SceneAnalysisOutput, error)
}

// ============================================
// STORAGE
// ============================================

// StorageClient defines the interface for file storage operations
type StorageClient interface {
	// GenerateUploadURL creates a presigned URL for uploading
	GenerateUploadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error)

	// GenerateDownloadURL creates a presigned URL for downloading
	GenerateDownloadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error)

	// Download fetches file content from storage
	Download(ctx context.Context, bucket, key string) (data []byte, contentType string, err error)

	// Upload stores file content to storage
	Upload(ctx context.Context, bucket, key string, data []byte, contentType string) error

	// Delete removes a file from storage
	Delete(ctx context.Context, bucket, key string) error
}
