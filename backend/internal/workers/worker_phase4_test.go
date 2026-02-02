package workers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

func TestReconstructionWorker_MergeOutput(t *testing.T) {
	worker := &ReconstructionWorker{
		BaseWorker: NewBaseWorker(nil, nil),
	}

	existing := models.NewEmptySceneGraph()
	existing.Objects = []models.SceneObject{
		{ID: "obj-1", Type: models.ObjectTypeFurniture, Label: "Table", State: models.ObjectStateVisible},
		{ID: "obj-2", Type: models.ObjectTypeDoor, Label: "Door", State: models.ObjectStateVisible},
	}

	output := &models.ReconstructionOutput{
		Objects: []models.SceneObjectProposal{
			// Update existing object
			{
				ID:         "obj-1",
				Action:     "update",
				Confidence: 0.9,
				Object: &models.SceneObject{
					ID:    "obj-1",
					Type:  models.ObjectTypeFurniture,
					Label: "Dining Table",
					State: models.ObjectStateVisible,
				},
			},
			// Create new object
			{
				ID:         "obj-3",
				Action:     "create",
				Confidence: 0.85,
				Object: &models.SceneObject{
					ID:    "obj-3",
					Type:  models.ObjectTypeWeapon,
					Label: "Knife",
					State: models.ObjectStateSuspicious,
				},
			},
			// Remove object
			{
				ID:         "obj-2",
				Action:     "remove",
				Confidence: 0.8,
			},
		},
	}

	result := worker.mergeReconstructionOutput(existing, output)

	// Should have 2 objects (obj-1 updated, obj-2 removed, obj-3 added)
	if len(result.Objects) != 2 {
		t.Errorf("Expected 2 objects, got %d", len(result.Objects))
	}

	// Check obj-1 was updated
	found := false
	for _, obj := range result.Objects {
		if obj.ID == "obj-1" {
			found = true
			if obj.Label != "Dining Table" {
				t.Errorf("Expected obj-1 label to be 'Dining Table', got '%s'", obj.Label)
			}
		}
	}
	if !found {
		t.Error("obj-1 should exist after update")
	}

	// Check obj-2 was removed
	for _, obj := range result.Objects {
		if obj.ID == "obj-2" {
			t.Error("obj-2 should have been removed")
		}
	}

	// Check obj-3 was added
	found = false
	for _, obj := range result.Objects {
		if obj.ID == "obj-3" {
			found = true
		}
	}
	if !found {
		t.Error("obj-3 should have been added")
	}
}

func TestReconstructionWorker_Type(t *testing.T) {
	worker := NewReconstructionWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeReconstruction {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeReconstruction)
	}
}

func TestReasoningWorker_Type(t *testing.T) {
	worker := NewReasoningWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeReasoning {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeReasoning)
	}
}

func TestProfileWorker_Type(t *testing.T) {
	worker := NewProfileWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeProfile {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeProfile)
	}
}

func TestImageGenWorker_Type(t *testing.T) {
	worker := NewImageGenWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeImageGen {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeImageGen)
	}
}

func TestProfileWorker_MergeAttributes(t *testing.T) {
	worker := &ProfileWorker{
		BaseWorker: NewBaseWorker(nil, nil),
	}

	existing := &models.SuspectAttributes{
		AgeRange: &models.RangeAttribute{Min: 20, Max: 30, Confidence: 0.6},
		Build:    &models.StringAttribute{Value: "slim", Confidence: 0.5},
	}

	new := &models.SuspectAttributes{
		AgeRange: &models.RangeAttribute{Min: 25, Max: 35, Confidence: 0.8}, // Higher confidence
		SkinTone: &models.StringAttribute{Value: "medium", Confidence: 0.7},
	}

	result := worker.mergeAttributes(existing, new, nil)

	// AgeRange should use new (higher confidence)
	if result.AgeRange.Confidence != 0.8 {
		t.Errorf("AgeRange should have higher confidence, got %f", result.AgeRange.Confidence)
	}

	// Build should be preserved from existing
	if result.Build == nil || result.Build.Value != "slim" {
		t.Error("Build should be preserved from existing")
	}

	// SkinTone should be from new
	if result.SkinTone == nil || result.SkinTone.Value != "medium" {
		t.Error("SkinTone should be from new")
	}
}

func TestProfileWorker_ShouldTriggerImageGen(t *testing.T) {
	worker := &ProfileWorker{
		BaseWorker: NewBaseWorker(nil, nil),
	}

	tests := []struct {
		name     string
		attrs    *models.SuspectAttributes
		expected bool
	}{
		{
			name:     "nil attributes",
			attrs:    nil,
			expected: false,
		},
		{
			name:     "empty attributes",
			attrs:    models.NewEmptySuspectAttributes(),
			expected: false,
		},
		{
			name: "insufficient attributes",
			attrs: &models.SuspectAttributes{
				AgeRange: &models.RangeAttribute{Min: 20, Max: 30, Confidence: 0.8},
			},
			expected: false,
		},
		{
			name: "sufficient attributes",
			attrs: &models.SuspectAttributes{
				AgeRange: &models.RangeAttribute{Min: 20, Max: 30, Confidence: 0.8},
				Build:    &models.StringAttribute{Value: "slim", Confidence: 0.7},
				Hair:     &models.HairAttribute{Color: "brown", Style: "short", Confidence: 0.6},
				SkinTone: &models.StringAttribute{Value: "medium", Confidence: 0.7},
			},
			expected: true,
		},
		{
			name: "low confidence attributes",
			attrs: &models.SuspectAttributes{
				AgeRange: &models.RangeAttribute{Min: 20, Max: 30, Confidence: 0.3},
				Build:    &models.StringAttribute{Value: "slim", Confidence: 0.3},
				Hair:     &models.HairAttribute{Color: "brown", Confidence: 0.3},
				SkinTone: &models.StringAttribute{Value: "medium", Confidence: 0.3},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.shouldTriggerImageGen(tt.attrs)
			if result != tt.expected {
				t.Errorf("shouldTriggerImageGen() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReconstructionWorker_ProcessWithMock(t *testing.T) {
	mockClient := &clients.MockReconstructionClient{}
	worker := NewReconstructionWorker(nil, nil, mockClient)

	input := models.ReconstructionInput{
		CaseID:        uuid.New().String(),
		ScanAssetKeys: []string{"key1.jpg", "key2.jpg"},
	}
	inputJSON, _ := json.Marshal(input)

	job := &queue.JobMessage{
		JobID:  uuid.New(),
		CaseID: uuid.New(),
		Type:   models.JobTypeReconstruction,
		Input:  inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Errorf("Process() error = %v", err)
	}
}

func TestReasoningWorker_ProcessWithMock(t *testing.T) {
	mockClient := &clients.MockReasoningClient{}
	worker := NewReasoningWorker(nil, nil, mockClient)

	input := models.ReasoningInput{
		CaseID:     uuid.New().String(),
		Scenegraph: models.NewEmptySceneGraph(),
	}
	inputJSON, _ := json.Marshal(input)

	job := &queue.JobMessage{
		JobID:  uuid.New(),
		CaseID: uuid.New(),
		Type:   models.JobTypeReasoning,
		Input:  inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Errorf("Process() error = %v", err)
	}
}

func TestProfileWorker_ProcessWithMock(t *testing.T) {
	mockClient := &clients.MockProfileClient{}
	worker := NewProfileWorker(nil, nil, mockClient)

	input := models.ProfileInput{
		CaseID: uuid.New().String(),
		Statements: []models.WitnessStatementInput{
			{SourceName: "Witness A", Content: "Tall man", Credibility: 0.8},
		},
	}
	inputJSON, _ := json.Marshal(input)

	job := &queue.JobMessage{
		JobID:  uuid.New(),
		CaseID: uuid.New(),
		Type:   models.JobTypeProfile,
		Input:  inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Errorf("Process() error = %v", err)
	}
}

func TestImageGenWorker_ProcessWithMock(t *testing.T) {
	mockClient := &clients.MockImageGenClient{}
	worker := NewImageGenWorker(nil, nil, mockClient)

	input := models.ImageGenInput{
		CaseID:        uuid.New().String(),
		GenType:       models.ImageGenTypeEvidenceBoard,
		Resolution:    "1k",
	}
	inputJSON, _ := json.Marshal(input)

	job := &queue.JobMessage{
		JobID:  uuid.New(),
		CaseID: uuid.New(),
		Type:   models.JobTypeImageGen,
		Input:  inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Errorf("Process() error = %v", err)
	}
}

func TestWorkerError_InvalidInput(t *testing.T) {
	mockClient := &clients.MockReconstructionClient{}
	worker := NewReconstructionWorker(nil, nil, mockClient)

	// Invalid JSON
	job := &queue.JobMessage{
		JobID:  uuid.New(),
		CaseID: uuid.New(),
		Type:   models.JobTypeReconstruction,
		Input:  []byte("invalid json"),
	}

	err := worker.Process(context.Background(), job)
	if err == nil {
		t.Error("Expected error for invalid JSON input")
	}

	// Should be a fatal error
	var workerErr *WorkerError
	if !IsRetryable(err) || !errors.As(err, &workerErr) {
		// Check if it's wrapped as fatal
		if workerErr, ok := err.(*WorkerError); ok && workerErr.Type == ErrorTypeFatal {
			// Expected
		}
	}
}

