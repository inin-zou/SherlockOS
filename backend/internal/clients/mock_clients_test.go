package clients

import (
	"context"
	"errors"
	"testing"

	"github.com/sherlockos/backend/internal/models"
)

func TestMockReconstructionClient_DefaultResponse(t *testing.T) {
	client := &MockReconstructionClient{}

	input := models.ReconstructionInput{
		CaseID:        "case_123",
		ScanAssetKeys: []string{"key1", "key2"},
	}

	output, err := client.Reconstruct(context.Background(), input)
	if err != nil {
		t.Fatalf("Reconstruct() error = %v", err)
	}

	if output == nil {
		t.Fatal("Reconstruct() output should not be nil")
	}

	if len(output.Objects) == 0 {
		t.Error("Reconstruct() should return at least one object")
	}

	if output.ProcessingStats.InputImages != 2 {
		t.Errorf("Reconstruct() InputImages = %v, want 2", output.ProcessingStats.InputImages)
	}
}

func TestMockReconstructionClient_CustomFunc(t *testing.T) {
	expectedErr := errors.New("model unavailable")

	client := &MockReconstructionClient{
		ReconstructFunc: func(ctx context.Context, input models.ReconstructionInput) (*models.ReconstructionOutput, error) {
			return nil, expectedErr
		},
	}

	_, err := client.Reconstruct(context.Background(), models.ReconstructionInput{})
	if err != expectedErr {
		t.Errorf("Reconstruct() error = %v, want %v", err, expectedErr)
	}
}

func TestMockReasoningClient_DefaultResponse(t *testing.T) {
	client := &MockReasoningClient{}

	input := models.ReasoningInput{
		CaseID:     "case_123",
		Scenegraph: models.NewEmptySceneGraph(),
	}

	output, err := client.Reason(context.Background(), input)
	if err != nil {
		t.Fatalf("Reason() error = %v", err)
	}

	if output == nil {
		t.Fatal("Reason() output should not be nil")
	}

	if len(output.Trajectories) == 0 {
		t.Error("Reason() should return at least one trajectory")
	}

	if output.Trajectories[0].Rank != 1 {
		t.Errorf("Reason() first trajectory rank = %v, want 1", output.Trajectories[0].Rank)
	}

	if len(output.NextStepSuggestions) == 0 {
		t.Error("Reason() should return at least one suggestion")
	}

	if output.ModelStats.ThinkingTokens == 0 {
		t.Error("Reason() should return non-zero thinking tokens")
	}
}

func TestMockReasoningClient_CustomFunc(t *testing.T) {
	customOutput := &models.ReasoningOutput{
		Trajectories: []models.Trajectory{
			{ID: "custom", Rank: 1, OverallConfidence: 0.99},
		},
	}

	client := &MockReasoningClient{
		ReasonFunc: func(ctx context.Context, input models.ReasoningInput) (*models.ReasoningOutput, error) {
			return customOutput, nil
		},
	}

	output, err := client.Reason(context.Background(), models.ReasoningInput{})
	if err != nil {
		t.Fatalf("Reason() error = %v", err)
	}

	if output.Trajectories[0].ID != "custom" {
		t.Error("Reason() should use custom function output")
	}
}

func TestMockImageGenClient_DefaultResponse(t *testing.T) {
	client := &MockImageGenClient{}

	tests := []struct {
		resolution    string
		expectedModel string
		expectedCost  float64
	}{
		{"1k", "nano-banana", 0.04},
		{"2k", "nano-banana-pro", 0.134},
		{"4k", "nano-banana-pro", 0.24},
	}

	for _, tt := range tests {
		t.Run(tt.resolution, func(t *testing.T) {
			input := ImageGenInput{
				CaseID:     "case_123",
				GenType:    "portrait",
				Resolution: tt.resolution,
			}

			output, err := client.Generate(context.Background(), input)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if output.ModelUsed != tt.expectedModel {
				t.Errorf("Generate() ModelUsed = %v, want %v", output.ModelUsed, tt.expectedModel)
			}

			if output.CostUSD != tt.expectedCost {
				t.Errorf("Generate() CostUSD = %v, want %v", output.CostUSD, tt.expectedCost)
			}

			if output.AssetKey == "" {
				t.Error("Generate() should return non-empty AssetKey")
			}

			if output.ThumbnailKey == "" {
				t.Error("Generate() should return non-empty ThumbnailKey")
			}
		})
	}
}

func TestMockImageGenClient_CustomFunc(t *testing.T) {
	customOutput := &ImageGenOutput{
		AssetKey:  "custom/key.png",
		ModelUsed: "custom-model",
		CostUSD:   0.5,
	}

	client := &MockImageGenClient{
		GenerateFunc: func(ctx context.Context, input ImageGenInput) (*ImageGenOutput, error) {
			return customOutput, nil
		},
	}

	output, err := client.Generate(context.Background(), ImageGenInput{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if output.AssetKey != "custom/key.png" {
		t.Error("Generate() should use custom function output")
	}
}

func TestMockProfileClient_DefaultResponse(t *testing.T) {
	client := &MockProfileClient{}

	statements := []models.WitnessStatementInput{
		{SourceName: "Witness A", Content: "Tall person", Credibility: 0.8},
	}

	attrs, err := client.ExtractProfile(context.Background(), statements, nil)
	if err != nil {
		t.Fatalf("ExtractProfile() error = %v", err)
	}

	if attrs == nil {
		t.Fatal("ExtractProfile() should return non-nil attributes")
	}

	if attrs.AgeRange == nil {
		t.Error("ExtractProfile() should set AgeRange")
	}

	if attrs.HeightRangeCm == nil {
		t.Error("ExtractProfile() should set HeightRangeCm")
	}

	if attrs.Build == nil {
		t.Error("ExtractProfile() should set Build")
	}
}

func TestMockProfileClient_CustomFunc(t *testing.T) {
	customAttrs := &models.SuspectAttributes{
		AgeRange: &models.RangeAttribute{
			Min:        40,
			Max:        50,
			Confidence: 0.95,
		},
	}

	client := &MockProfileClient{
		ExtractProfileFunc: func(ctx context.Context, statements []models.WitnessStatementInput, existing *models.SuspectAttributes) (*models.SuspectAttributes, error) {
			return customAttrs, nil
		},
	}

	attrs, err := client.ExtractProfile(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ExtractProfile() error = %v", err)
	}

	if attrs.AgeRange.Min != 40 {
		t.Error("ExtractProfile() should use custom function output")
	}
}

// Test that mock clients implement interfaces
func TestMockClients_ImplementInterfaces(t *testing.T) {
	// These will fail to compile if interfaces are not implemented
	var _ ReconstructionClient = &MockReconstructionClient{}
	var _ ReasoningClient = &MockReasoningClient{}
	var _ ImageGenClient = &MockImageGenClient{}
	var _ ProfileClient = &MockProfileClient{}
}
