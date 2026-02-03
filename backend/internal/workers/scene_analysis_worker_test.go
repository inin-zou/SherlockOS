package workers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

func TestSceneAnalysisWorker_Type(t *testing.T) {
	worker := NewSceneAnalysisWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeSceneAnalysis {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeSceneAnalysis)
	}
}

func TestSceneAnalysisWorker_ProcessWithMock(t *testing.T) {
	mockClient := &clients.MockSceneAnalysisClient{}
	worker := NewSceneAnalysisWorker(nil, nil, mockClient)

	input := models.SceneAnalysisInput{
		CaseID:    uuid.New().String(),
		ImageKeys: []string{"cases/test/scans/image1.jpg"},
		Mode:      "full_analysis",
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeSceneAnalysis,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}

func TestSceneAnalysisWorker_InvalidInput(t *testing.T) {
	mockClient := &clients.MockSceneAnalysisClient{}
	worker := NewSceneAnalysisWorker(nil, nil, mockClient)

	// Invalid JSON input
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeSceneAnalysis,
		Input: []byte("invalid json"),
	}

	err := worker.Process(context.Background(), job)
	if err == nil {
		t.Error("Process() expected error for invalid input")
	}

	// Check it's a fatal error (not retryable)
	if we, ok := err.(*WorkerError); ok {
		if we.Type != ErrorTypeFatal {
			t.Errorf("Process() error type = %v, want %v", we.Type, ErrorTypeFatal)
		}
	}
}

func TestSceneAnalysisWorker_MissingImageKeys(t *testing.T) {
	mockClient := &clients.MockSceneAnalysisClient{}
	worker := NewSceneAnalysisWorker(nil, nil, mockClient)

	// Missing required image keys
	input := models.SceneAnalysisInput{
		CaseID:    uuid.New().String(),
		ImageKeys: []string{}, // Empty
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeSceneAnalysis,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err == nil {
		t.Error("Process() expected error for missing image keys")
	}
}

func TestSceneAnalysisWorker_OutputContainsObjects(t *testing.T) {
	mockClient := &clients.MockSceneAnalysisClient{}
	worker := NewSceneAnalysisWorker(nil, nil, mockClient)

	input := models.SceneAnalysisInput{
		CaseID:    uuid.New().String(),
		ImageKeys: []string{"cases/test/scans/image1.jpg"},
		Mode:      "full_analysis",
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeSceneAnalysis,
		Input: inputJSON,
	}

	// Process should complete without error
	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Mock client returns objects - verify expected behavior
	// (actual verification would require database access in integration tests)
}

func TestSceneAnalysisWorker_ModeSetting(t *testing.T) {
	mockClient := &clients.MockSceneAnalysisClient{}
	worker := NewSceneAnalysisWorker(nil, nil, mockClient)

	tests := []struct {
		mode string
	}{
		{"object_detection"},
		{"evidence_search"},
		{"full_analysis"},
		{""},  // Default mode
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			input := models.SceneAnalysisInput{
				CaseID:    uuid.New().String(),
				ImageKeys: []string{"test.jpg"},
				Mode:      tt.mode,
			}

			inputJSON, _ := json.Marshal(input)
			job := &queue.JobMessage{
				JobID: uuid.New(),
				Type:  models.JobTypeSceneAnalysis,
				Input: inputJSON,
			}

			err := worker.Process(context.Background(), job)
			if err != nil {
				t.Errorf("Process() with mode %q failed: %v", tt.mode, err)
			}
		})
	}
}
