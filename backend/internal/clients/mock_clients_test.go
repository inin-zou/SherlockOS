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
		{"1k", "gemini-2.5-flash-image", 0.04},      // Nano Banana
		{"2k", "gemini-3-pro-image-preview", 0.134}, // Nano Banana Pro
		{"4k", "gemini-3-pro-image-preview", 0.24},  // Nano Banana Pro
	}

	for _, tt := range tests {
		t.Run(tt.resolution, func(t *testing.T) {
			input := models.ImageGenInput{
				CaseID:     "case_123",
				GenType:    models.ImageGenTypePortrait,
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
	customOutput := &models.ImageGenOutput{
		AssetKey:  "custom/key.png",
		ModelUsed: "custom-model",
		CostUSD:   0.5,
	}

	client := &MockImageGenClient{
		GenerateFunc: func(ctx context.Context, input models.ImageGenInput) (*models.ImageGenOutput, error) {
			return customOutput, nil
		},
	}

	output, err := client.Generate(context.Background(), models.ImageGenInput{})
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
	var _ ReplayClient = &MockReplayClient{}
	var _ Asset3DClient = &MockAsset3DClient{}
	var _ SceneAnalysisClient = &MockSceneAnalysisClient{}
}

// ============================================
// REPLAY CLIENT TESTS
// ============================================

func TestMockReplayClient_DefaultResponse(t *testing.T) {
	client := &MockReplayClient{}

	input := models.ReplayInput{
		CaseID:       "case_123",
		TrajectoryID: "traj_001",
		Perspective:  "third_person",
		FrameCount:   125,
		Resolution:   "480p",
	}

	output, err := client.GenerateReplay(context.Background(), input)
	if err != nil {
		t.Fatalf("GenerateReplay() error = %v", err)
	}

	if output == nil {
		t.Fatal("GenerateReplay() output should not be nil")
	}

	if output.VideoAssetKey == "" {
		t.Error("GenerateReplay() should return non-empty VideoAssetKey")
	}

	if output.FPS != 24 {
		t.Errorf("GenerateReplay() FPS = %v, want 24", output.FPS)
	}

	if output.ModelUsed != "hy-world-1.5" {
		t.Errorf("GenerateReplay() ModelUsed = %v, want hy-world-1.5", output.ModelUsed)
	}

	if output.FrameCount != 125 {
		t.Errorf("GenerateReplay() FrameCount = %v, want 125", output.FrameCount)
	}
}

func TestMockReplayClient_CustomFunc(t *testing.T) {
	customOutput := &models.ReplayOutput{
		VideoAssetKey: "custom/video.mp4",
		ModelUsed:     "custom-model",
	}

	client := &MockReplayClient{
		GenerateReplayFunc: func(ctx context.Context, input models.ReplayInput) (*models.ReplayOutput, error) {
			return customOutput, nil
		},
	}

	output, err := client.GenerateReplay(context.Background(), models.ReplayInput{CaseID: "test"})
	if err != nil {
		t.Fatalf("GenerateReplay() error = %v", err)
	}

	if output.VideoAssetKey != "custom/video.mp4" {
		t.Error("GenerateReplay() should use custom function output")
	}
}

// ============================================
// ASSET 3D CLIENT TESTS
// ============================================

func TestMockAsset3DClient_DefaultResponse(t *testing.T) {
	client := &MockAsset3DClient{}

	input := models.Asset3DInput{
		CaseID:       "case_123",
		ImageKey:     "cases/case_123/scans/evidence.jpg",
		ItemType:     "weapon",
		WithTexture:  true,
		OutputFormat: "glb",
	}

	output, err := client.Generate3DAsset(context.Background(), input)
	if err != nil {
		t.Fatalf("Generate3DAsset() error = %v", err)
	}

	if output == nil {
		t.Fatal("Generate3DAsset() output should not be nil")
	}

	if output.MeshAssetKey == "" {
		t.Error("Generate3DAsset() should return non-empty MeshAssetKey")
	}

	if output.Format != "glb" {
		t.Errorf("Generate3DAsset() Format = %v, want glb", output.Format)
	}

	if output.ModelUsed != "hunyuan3d-2" {
		t.Errorf("Generate3DAsset() ModelUsed = %v, want hunyuan3d-2", output.ModelUsed)
	}

	if !output.HasTexture {
		t.Error("Generate3DAsset() HasTexture should be true")
	}
}

func TestMockAsset3DClient_CustomFunc(t *testing.T) {
	customOutput := &models.Asset3DOutput{
		MeshAssetKey: "custom/mesh.obj",
		Format:       "obj",
		ModelUsed:    "custom-model",
	}

	client := &MockAsset3DClient{
		Generate3DAssetFunc: func(ctx context.Context, input models.Asset3DInput) (*models.Asset3DOutput, error) {
			return customOutput, nil
		},
	}

	output, err := client.Generate3DAsset(context.Background(), models.Asset3DInput{CaseID: "test", ImageKey: "test.jpg"})
	if err != nil {
		t.Fatalf("Generate3DAsset() error = %v", err)
	}

	if output.MeshAssetKey != "custom/mesh.obj" {
		t.Error("Generate3DAsset() should use custom function output")
	}
}

// ============================================
// SCENE ANALYSIS CLIENT TESTS
// ============================================

func TestMockSceneAnalysisClient_DefaultResponse(t *testing.T) {
	client := &MockSceneAnalysisClient{}

	input := models.SceneAnalysisInput{
		CaseID:    "case_123",
		ImageKeys: []string{"cases/case_123/scans/scene1.jpg"},
		Mode:      "full_analysis",
	}

	output, err := client.AnalyzeScene(context.Background(), input)
	if err != nil {
		t.Fatalf("AnalyzeScene() error = %v", err)
	}

	if output == nil {
		t.Fatal("AnalyzeScene() output should not be nil")
	}

	if len(output.DetectedObjects) == 0 {
		t.Error("AnalyzeScene() should return detected objects")
	}

	if len(output.PotentialEvidence) == 0 {
		t.Error("AnalyzeScene() should return potential evidence")
	}

	if output.SceneDescription == "" {
		t.Error("AnalyzeScene() should return scene description")
	}

	if output.ModelUsed != "gemini-2.0-flash-001" {
		t.Errorf("AnalyzeScene() ModelUsed = %v, want gemini-2.0-flash-001", output.ModelUsed)
	}

	// Check that suspicious objects are marked
	hasSuspicious := false
	for _, obj := range output.DetectedObjects {
		if obj.IsSuspicious {
			hasSuspicious = true
			break
		}
	}
	if !hasSuspicious {
		t.Error("AnalyzeScene() should return at least one suspicious object")
	}
}

func TestMockSceneAnalysisClient_CustomFunc(t *testing.T) {
	customOutput := &models.SceneAnalysisOutput{
		DetectedObjects: []models.DetectedObject{
			{ID: "custom_obj", Type: "weapon", Label: "Knife"},
		},
		ModelUsed: "custom-model",
	}

	client := &MockSceneAnalysisClient{
		AnalyzeSceneFunc: func(ctx context.Context, input models.SceneAnalysisInput) (*models.SceneAnalysisOutput, error) {
			return customOutput, nil
		},
	}

	output, err := client.AnalyzeScene(context.Background(), models.SceneAnalysisInput{CaseID: "test", ImageKeys: []string{"test.jpg"}})
	if err != nil {
		t.Fatalf("AnalyzeScene() error = %v", err)
	}

	if output.DetectedObjects[0].ID != "custom_obj" {
		t.Error("AnalyzeScene() should use custom function output")
	}
}

// ============================================
// STORAGE CLIENT TESTS
// ============================================

func TestMockStorageClient_DefaultResponse(t *testing.T) {
	client := &MockStorageClient{}

	// Test GenerateUploadURL
	uploadURL, err := client.GenerateUploadURL(context.Background(), "assets", "test/file.jpg", 3600)
	if err != nil {
		t.Fatalf("GenerateUploadURL() error = %v", err)
	}
	if uploadURL == "" {
		t.Error("GenerateUploadURL() should return non-empty URL")
	}

	// Test GenerateDownloadURL
	downloadURL, err := client.GenerateDownloadURL(context.Background(), "assets", "test/file.jpg", 3600)
	if err != nil {
		t.Fatalf("GenerateDownloadURL() error = %v", err)
	}
	if downloadURL == "" {
		t.Error("GenerateDownloadURL() should return non-empty URL")
	}

	// Test Download
	data, contentType, err := client.Download(context.Background(), "assets", "test/file.jpg")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("Download() should return non-empty data")
	}
	if contentType != "image/png" {
		t.Errorf("Download() contentType = %v, want image/png", contentType)
	}

	// Test Upload
	err = client.Upload(context.Background(), "assets", "test/file.jpg", []byte("test data"), "image/jpeg")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	// Test Delete
	err = client.Delete(context.Background(), "assets", "test/file.jpg")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestMockStorageClient_CustomFunc(t *testing.T) {
	customData := []byte("custom image data")
	customContentType := "image/jpeg"

	client := &MockStorageClient{
		DownloadFunc: func(ctx context.Context, bucket, key string) ([]byte, string, error) {
			return customData, customContentType, nil
		},
	}

	data, contentType, err := client.Download(context.Background(), "assets", "test/file.jpg")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	if string(data) != string(customData) {
		t.Error("Download() should use custom function output")
	}
	if contentType != customContentType {
		t.Errorf("Download() contentType = %v, want %v", contentType, customContentType)
	}
}

// Test that mock clients implement interfaces
func TestMockStorageClient_ImplementsInterface(t *testing.T) {
	var _ StorageClient = &MockStorageClient{}
}
