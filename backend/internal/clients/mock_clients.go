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
	GenerateFunc func(ctx context.Context, input ImageGenInput) (*ImageGenOutput, error)
}

// Generate implements ImageGenClient
func (m *MockImageGenClient) Generate(ctx context.Context, input ImageGenInput) (*ImageGenOutput, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, input)
	}

	// Determine model based on resolution
	modelUsed := "nano-banana"
	cost := 0.04
	width, height := 1024, 1024

	if input.Resolution == "2k" {
		modelUsed = "nano-banana-pro"
		cost = 0.134
		width, height = 2048, 2048
	} else if input.Resolution == "4k" {
		modelUsed = "nano-banana-pro"
		cost = 0.24
		width, height = 4096, 4096
	}

	return &ImageGenOutput{
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
