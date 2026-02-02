package clients

import (
	"context"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/models"
)

// MockReconstructionClient is a mock implementation for testing
type MockReconstructionClient struct {
	ReconstructFunc func(ctx context.Context, input models.ReconstructionInput) (*models.ReconstructionOutput, error)
}

// Reconstruct implements ReconstructionClient
func (m *MockReconstructionClient) Reconstruct(ctx context.Context, input models.ReconstructionInput) (*models.ReconstructionOutput, error) {
	if m.ReconstructFunc != nil {
		return m.ReconstructFunc(ctx, input)
	}

	// Default mock response
	return &models.ReconstructionOutput{
		Objects: []models.SceneObjectProposal{
			{
				ID:         uuid.New().String(),
				Action:     "create",
				Confidence: 0.85,
				Object: &models.SceneObject{
					ID:         uuid.New().String(),
					Type:       models.ObjectTypeFurniture,
					Label:      "Table",
					Pose:       models.NewDefaultPose(),
					BBox:       models.BoundingBox{Min: [3]float64{0, 0, 0}, Max: [3]float64{1, 1, 1}},
					State:      models.ObjectStateVisible,
					Confidence: 0.85,
				},
				SourceImages: input.ScanAssetKeys,
			},
		},
		UncertaintyRegions: []models.UncertaintyRegion{},
		ProcessingStats: models.ProcessingStats{
			InputImages:      len(input.ScanAssetKeys),
			DetectedObjects:  1,
			ProcessingTimeMs: 1500,
		},
	}, nil
}

// MockReasoningClient is a mock implementation for testing
type MockReasoningClient struct {
	ReasonFunc func(ctx context.Context, input models.ReasoningInput) (*models.ReasoningOutput, error)
}

// Reason implements ReasoningClient
func (m *MockReasoningClient) Reason(ctx context.Context, input models.ReasoningInput) (*models.ReasoningOutput, error) {
	if m.ReasonFunc != nil {
		return m.ReasonFunc(ctx, input)
	}

	// Default mock response
	return &models.ReasoningOutput{
		Trajectories: []models.Trajectory{
			{
				ID:                uuid.New().String(),
				Rank:              1,
				OverallConfidence: 0.75,
				Segments: []models.TrajectorySegment{
					{
						ID:           uuid.New().String(),
						FromPosition: [3]float64{0, 0, 0},
						ToPosition:   [3]float64{5, 0, 3},
						Confidence:   0.8,
						Explanation:  "Entry through main door based on lock damage evidence",
						EvidenceRefs: []models.EvidenceRef{
							{
								EvidenceID: "ev_001",
								Relevance:  "supports",
								Weight:     0.9,
							},
						},
					},
				},
			},
		},
		UncertaintyAreas: []models.UncertaintyRegion{},
		NextStepSuggestions: []models.Suggestion{
			{
				Type:        "collect_evidence",
				Description: "Check for fingerprints on door handle",
				Priority:    "high",
			},
		},
		ThinkingSummary: "Analysis based on evidence positions and physical constraints",
		ModelStats: models.ModelStats{
			ThinkingTokens: 4096,
			OutputTokens:   1024,
			LatencyMs:      2500,
		},
	}, nil
}

// MockImageGenClient is a mock implementation for testing
type MockImageGenClient struct {
	GenerateFunc func(ctx context.Context, input models.ImageGenInput) (*models.ImageGenOutput, error)
}

// Generate implements ImageGenClient
func (m *MockImageGenClient) Generate(ctx context.Context, input models.ImageGenInput) (*models.ImageGenOutput, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, input)
	}

	// Determine model based on resolution (Nano Banana series)
	modelUsed := "gemini-2.5-flash-image" // Nano Banana
	cost := 0.04
	width, height := 1024, 1024

	if input.Resolution == "2k" {
		modelUsed = "gemini-3-pro-image-preview" // Nano Banana Pro
		cost = 0.134
		width, height = 2048, 2048
	} else if input.Resolution == "4k" {
		modelUsed = "gemini-3-pro-image-preview" // Nano Banana Pro
		cost = 0.24
		width, height = 4096, 4096
	}

	return &models.ImageGenOutput{
		AssetKey:       "cases/" + input.CaseID + "/generated/" + uuid.New().String() + ".png",
		ThumbnailKey:   "cases/" + input.CaseID + "/generated/" + uuid.New().String() + "_thumb.png",
		Width:          width,
		Height:         height,
		ModelUsed:      modelUsed,
		GenerationTime: 3000,
		CostUSD:        cost,
	}, nil
}

// MockProfileClient is a mock implementation for testing
type MockProfileClient struct {
	ExtractProfileFunc func(ctx context.Context, statements []models.WitnessStatementInput, existing *models.SuspectAttributes) (*models.SuspectAttributes, error)
}

// ExtractProfile implements ProfileClient
func (m *MockProfileClient) ExtractProfile(ctx context.Context, statements []models.WitnessStatementInput, existing *models.SuspectAttributes) (*models.SuspectAttributes, error) {
	if m.ExtractProfileFunc != nil {
		return m.ExtractProfileFunc(ctx, statements, existing)
	}

	// Default mock response with sample attributes
	attrs := models.NewEmptySuspectAttributes()
	attrs.AgeRange = &models.RangeAttribute{
		Min:        25,
		Max:        35,
		Confidence: 0.7,
	}
	attrs.HeightRangeCm = &models.RangeAttribute{
		Min:        170,
		Max:        180,
		Confidence: 0.8,
	}
	attrs.Build = &models.StringAttribute{
		Value:      "average",
		Confidence: 0.6,
	}

	return attrs, nil
}

// ============================================
// REPLAY CLIENT (HY-World-1.5)
// ============================================

// MockReplayClient is a mock implementation for testing
type MockReplayClient struct {
	GenerateReplayFunc func(ctx context.Context, input models.ReplayInput) (*models.ReplayOutput, error)
}

// GenerateReplay implements ReplayClient
func (m *MockReplayClient) GenerateReplay(ctx context.Context, input models.ReplayInput) (*models.ReplayOutput, error) {
	if m.GenerateReplayFunc != nil {
		return m.GenerateReplayFunc(ctx, input)
	}

	// Default mock response
	frameCount := input.FrameCount
	if frameCount == 0 {
		frameCount = 125
	}

	return &models.ReplayOutput{
		VideoAssetKey:  "cases/" + input.CaseID + "/replay/" + uuid.New().String() + ".mp4",
		ThumbnailKey:   "cases/" + input.CaseID + "/replay/" + uuid.New().String() + "_thumb.png",
		FrameCount:     frameCount,
		FPS:            24,
		DurationMs:     int64(frameCount * 1000 / 24),
		Resolution:     input.Resolution,
		ModelUsed:      "hy-world-1.5",
		GenerationTime: 15000, // ~15 seconds for mock
	}, nil
}

// ============================================
// ASSET 3D CLIENT (Hunyuan3D-2)
// ============================================

// MockAsset3DClient is a mock implementation for testing
type MockAsset3DClient struct {
	Generate3DAssetFunc func(ctx context.Context, input models.Asset3DInput) (*models.Asset3DOutput, error)
}

// Generate3DAsset implements Asset3DClient
func (m *MockAsset3DClient) Generate3DAsset(ctx context.Context, input models.Asset3DInput) (*models.Asset3DOutput, error) {
	if m.Generate3DAssetFunc != nil {
		return m.Generate3DAssetFunc(ctx, input)
	}

	format := input.OutputFormat
	if format == "" {
		format = "glb"
	}

	// Default mock response
	return &models.Asset3DOutput{
		MeshAssetKey:   "cases/" + input.CaseID + "/models/" + uuid.New().String() + "." + format,
		ThumbnailKey:   "cases/" + input.CaseID + "/models/" + uuid.New().String() + "_thumb.png",
		Format:         format,
		HasTexture:     input.WithTexture,
		VertexCount:    15000,
		ModelUsed:      "hunyuan3d-2",
		GenerationTime: 60000, // ~60 seconds for mock
	}, nil
}

// ============================================
// SCENE ANALYSIS CLIENT (Gemini Vision)
// ============================================

// MockSceneAnalysisClient is a mock implementation for testing
type MockSceneAnalysisClient struct {
	AnalyzeSceneFunc func(ctx context.Context, input models.SceneAnalysisInput) (*models.SceneAnalysisOutput, error)
}

// AnalyzeScene implements SceneAnalysisClient
func (m *MockSceneAnalysisClient) AnalyzeScene(ctx context.Context, input models.SceneAnalysisInput) (*models.SceneAnalysisOutput, error) {
	if m.AnalyzeSceneFunc != nil {
		return m.AnalyzeSceneFunc(ctx, input)
	}

	// Default mock response with sample detected objects
	return &models.SceneAnalysisOutput{
		DetectedObjects: []models.DetectedObject{
			{
				ID:                  uuid.New().String(),
				Type:                "furniture",
				Label:               "Wooden Table",
				PositionDescription: "center of room",
				Confidence:          0.92,
				IsSuspicious:        false,
				SourceImageKey:      input.ImageKeys[0],
			},
			{
				ID:                  uuid.New().String(),
				Type:                "evidence_item",
				Label:               "Broken Glass",
				PositionDescription: "near window, floor level",
				Confidence:          0.87,
				IsSuspicious:        true,
				Notes:               "Possible point of entry",
				SourceImageKey:      input.ImageKeys[0],
			},
		},
		PotentialEvidence: []string{
			"Broken glass fragments near window",
			"Disturbed furniture arrangement",
		},
		SceneDescription: "Indoor room with signs of forced entry through window",
		Anomalies:        []string{"Window glass broken from outside"},
		AnalysisTime:     2500,
		ModelUsed:        "gemini-2.0-flash-001", // Fast, accurate vision model for scene analysis
	}, nil
}

// ============================================
// STORAGE CLIENT (Supabase Storage)
// ============================================

// MockStorageClient is a mock implementation for testing
type MockStorageClient struct {
	GenerateUploadURLFunc   func(ctx context.Context, bucket, key string, expiresIn int) (string, error)
	GenerateDownloadURLFunc func(ctx context.Context, bucket, key string, expiresIn int) (string, error)
	DownloadFunc            func(ctx context.Context, bucket, key string) ([]byte, string, error)
	UploadFunc              func(ctx context.Context, bucket, key string, data []byte, contentType string) error
	DeleteFunc              func(ctx context.Context, bucket, key string) error
}

// GenerateUploadURL implements StorageClient
func (m *MockStorageClient) GenerateUploadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error) {
	if m.GenerateUploadURLFunc != nil {
		return m.GenerateUploadURLFunc(ctx, bucket, key, expiresIn)
	}
	return "https://example.supabase.co/storage/v1/object/sign/" + bucket + "/" + key + "?token=mock", nil
}

// GenerateDownloadURL implements StorageClient
func (m *MockStorageClient) GenerateDownloadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error) {
	if m.GenerateDownloadURLFunc != nil {
		return m.GenerateDownloadURLFunc(ctx, bucket, key, expiresIn)
	}
	return "https://example.supabase.co/storage/v1/object/sign/" + bucket + "/" + key + "?token=mock", nil
}

// Download implements StorageClient
func (m *MockStorageClient) Download(ctx context.Context, bucket, key string) ([]byte, string, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, bucket, key)
	}
	// Return a minimal 1x1 white PNG for testing
	mockPNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
		0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59, 0xE7, 0x00, 0x00, 0x00,
		0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	return mockPNG, "image/png", nil
}

// Upload implements StorageClient
func (m *MockStorageClient) Upload(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, bucket, key, data, contentType)
	}
	return nil
}

// Delete implements StorageClient
func (m *MockStorageClient) Delete(ctx context.Context, bucket, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, bucket, key)
	}
	return nil
}
