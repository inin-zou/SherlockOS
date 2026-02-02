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

func TestAsset3DWorker_Type(t *testing.T) {
	worker := NewAsset3DWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeAsset3D {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeAsset3D)
	}
}

func TestAsset3DWorker_ProcessWithMock(t *testing.T) {
	mockClient := &clients.MockAsset3DClient{}
	worker := NewAsset3DWorker(nil, nil, mockClient)

	input := models.Asset3DInput{
		CaseID:       uuid.New().String(),
		ImageKey:     "cases/test/scans/evidence.jpg",
		ItemType:     "weapon",
		WithTexture:  true,
		OutputFormat: "glb",
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeAsset3D,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}

func TestAsset3DWorker_InvalidInput(t *testing.T) {
	mockClient := &clients.MockAsset3DClient{}
	worker := NewAsset3DWorker(nil, nil, mockClient)

	// Invalid JSON input
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeAsset3D,
		Input: []byte("invalid json"),
	}

	err := worker.Process(context.Background(), job)
	if err == nil {
		t.Error("Process() expected error for invalid input")
	}

	// Check it's a fatal error
	if we, ok := err.(*WorkerError); ok {
		if we.Type != ErrorTypeFatal {
			t.Errorf("Process() error type = %v, want %v", we.Type, ErrorTypeFatal)
		}
	}
}

func TestAsset3DWorker_MissingImageKey(t *testing.T) {
	mockClient := &clients.MockAsset3DClient{}
	worker := NewAsset3DWorker(nil, nil, mockClient)

	// Missing required image key
	input := models.Asset3DInput{
		CaseID:   uuid.New().String(),
		ImageKey: "", // Empty
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeAsset3D,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err == nil {
		t.Error("Process() expected error for missing image key")
	}
}

func TestAsset3DWorker_CustomOutput(t *testing.T) {
	customOutput := &models.Asset3DOutput{
		MeshAssetKey:   "custom/mesh.glb",
		Format:         "glb",
		HasTexture:     true,
		VertexCount:    20000,
		ModelUsed:      "hunyuan3d-2",
		GenerationTime: 45000,
	}

	mockClient := &clients.MockAsset3DClient{
		Generate3DAssetFunc: func(ctx context.Context, input models.Asset3DInput) (*models.Asset3DOutput, error) {
			return customOutput, nil
		},
	}

	worker := NewAsset3DWorker(nil, nil, mockClient)

	input := models.Asset3DInput{
		CaseID:   uuid.New().String(),
		ImageKey: "cases/test/scans/evidence.jpg",
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeAsset3D,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}
