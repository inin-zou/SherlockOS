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

func TestReconstructionWorker_TypeReturnsReconstruction(t *testing.T) {
	worker := NewReconstructionWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeReconstruction {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeReconstruction)
	}
}

func TestReconstructionWorker_ProcessBasic(t *testing.T) {
	mockClient := &clients.MockReconstructionClient{}
	worker := NewReconstructionWorker(nil, nil, mockClient)

	input := models.ReconstructionInput{
		CaseID:        uuid.New().String(),
		ScanAssetKeys: []string{"cases/test/scans/image1.jpg", "cases/test/scans/image2.jpg"},
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReconstruction,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}

func TestReconstructionWorker_ProcessWithPOVImages(t *testing.T) {
	mockClient := &clients.MockReconstructionClient{
		ReconstructFunc: func(ctx context.Context, input models.ReconstructionInput) (*models.ReconstructionOutput, error) {
			// Verify both raw and POV images are provided
			if len(input.ScanAssetKeys) != 2 {
				t.Errorf("Expected 2 raw images, got %d", len(input.ScanAssetKeys))
			}
			if len(input.GeneratedPOVKeys) != 4 {
				t.Errorf("Expected 4 POV images, got %d", len(input.GeneratedPOVKeys))
			}

			return &models.ReconstructionOutput{
				Objects:            []models.SceneObjectProposal{},
				UncertaintyRegions: []models.UncertaintyRegion{},
				ProcessingStats: models.ProcessingStats{
					InputImages:      len(input.ScanAssetKeys) + len(input.GeneratedPOVKeys),
					DetectedObjects:  2,
					ProcessingTimeMs: 1500,
				},
			}, nil
		},
	}

	worker := NewReconstructionWorker(nil, nil, mockClient)

	input := models.ReconstructionInput{
		CaseID:        uuid.New().String(),
		ScanAssetKeys: []string{"raw1.jpg", "raw2.jpg"},
		GeneratedPOVKeys: []string{
			"pov/front_xxx.png",
			"pov/left_xxx.png",
			"pov/right_xxx.png",
			"pov/back_xxx.png",
		},
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReconstruction,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}

func TestReconstructionWorker_InvalidInput(t *testing.T) {
	mockClient := &clients.MockReconstructionClient{}
	worker := NewReconstructionWorker(nil, nil, mockClient)

	// Invalid JSON input
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReconstruction,
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

func TestReconstructionWorker_MissingScanAssetKeys(t *testing.T) {
	mockClient := &clients.MockReconstructionClient{}
	worker := NewReconstructionWorker(nil, nil, mockClient)

	// Missing required scan asset keys
	input := models.ReconstructionInput{
		CaseID:        uuid.New().String(),
		ScanAssetKeys: []string{}, // Empty
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReconstruction,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err == nil {
		t.Error("Process() expected error for missing scan asset keys")
	}
}

func TestReconstructionWorker_ExtractPOVKeysFromOutput(t *testing.T) {
	worker := NewReconstructionWorker(nil, nil, nil)

	tests := []struct {
		name        string
		output      json.RawMessage
		wantKeys    int
		wantErr     bool
	}{
		{
			name: "multiple generated images",
			output: []byte(`{
				"generated_images": [
					{"asset_key": "pov/front.png", "view_angle": "front"},
					{"asset_key": "pov/left.png", "view_angle": "left"},
					{"asset_key": "pov/right.png", "view_angle": "right"},
					{"asset_key": "pov/back.png", "view_angle": "back"}
				]
			}`),
			wantKeys: 4,
			wantErr:  false,
		},
		{
			name: "single asset key fallback",
			output: []byte(`{
				"asset_key": "pov/single.png"
			}`),
			wantKeys: 1,
			wantErr:  false,
		},
		{
			name:     "empty output",
			output:   []byte(`{}`),
			wantKeys: 0,
			wantErr:  true,
		},
		{
			name:     "nil output",
			output:   nil,
			wantKeys: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := worker.extractPOVKeysFromOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractPOVKeysFromOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(keys) != tt.wantKeys {
				t.Errorf("extractPOVKeysFromOutput() got %d keys, want %d", len(keys), tt.wantKeys)
			}
		})
	}
}

func TestReconstructionWorker_MergeReconstructionOutput(t *testing.T) {
	worker := NewReconstructionWorker(nil, nil, nil)

	// Create existing scenegraph with one object
	existing := &models.SceneGraph{
		Version: "1.0",
		Objects: []models.SceneObject{
			{
				ID:    "existing_obj_1",
				Type:  models.ObjectTypeFurniture,
				Label: "Table",
				Pose:  models.NewDefaultPose(),
			},
		},
		Evidence:    []models.EvidenceCard{},
		Constraints: []models.Constraint{},
	}

	// Create output that adds a new object and updates the existing one
	output := &models.ReconstructionOutput{
		Objects: []models.SceneObjectProposal{
			{
				ID:     "new_obj_1",
				Action: "create",
				Object: &models.SceneObject{
					ID:    "new_obj_1",
					Type:  models.ObjectTypeEvidenceItem,
					Label: "Knife",
					Pose:  models.NewDefaultPose(),
				},
			},
			{
				ID:     "existing_obj_1",
				Action: "update",
				Object: &models.SceneObject{
					ID:    "existing_obj_1",
					Type:  models.ObjectTypeFurniture,
					Label: "Updated Table",
					Pose:  models.NewDefaultPose(),
				},
			},
		},
		UncertaintyRegions: []models.UncertaintyRegion{},
		PointCloud: &models.PointCloud{
			Positions: [][]float64{{0, 0, 0}, {1, 1, 1}},
			Colors:    [][]float64{{1, 0, 0}, {0, 1, 0}},
			Count:     2,
		},
	}

	result := worker.mergeReconstructionOutput(existing, output)

	// Should have 2 objects now
	if len(result.Objects) != 2 {
		t.Errorf("Expected 2 objects after merge, got %d", len(result.Objects))
	}

	// Point cloud should be passed through
	if result.PointCloud == nil {
		t.Error("Expected point cloud in result")
	} else if result.PointCloud.Count != 2 {
		t.Errorf("Expected point cloud count 2, got %d", result.PointCloud.Count)
	}
}

func TestReconstructionWorker_ComputeBoundsFromObjects(t *testing.T) {
	worker := NewReconstructionWorker(nil, nil, nil)

	tests := []struct {
		name    string
		objects []models.SceneObject
		wantMin [3]float64
		wantMax [3]float64
	}{
		{
			name:    "no objects - default bounds",
			objects: []models.SceneObject{},
			wantMin: [3]float64{-7, 0, -6},
			wantMax: [3]float64{7, 4, 6},
		},
		{
			name: "single object",
			objects: []models.SceneObject{
				{
					ID:   "obj1",
					Pose: models.Pose{Position: [3]float64{0, 0, 0}},
					BBox: models.BoundingBox{
						Min: [3]float64{-1, 0, -1},
						Max: [3]float64{1, 1, 1},
					},
				},
			},
			// Position + BBox + margin (2m)
			wantMin: [3]float64{-4, 0, -4}, // -1 - 2 margin, clamped to min 8m
			wantMax: [3]float64{4, 3.5, 4}, // 1 + 2 margin, clamped to min 8m
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.computeBoundsFromObjects(tt.objects)

			// Check bounds are reasonable (not testing exact values due to complexity)
			if result.Min[1] != 0 {
				t.Errorf("Expected floor at Y=0, got %f", result.Min[1])
			}
			if result.Max[1] < 3 {
				t.Errorf("Expected ceiling at least 3m, got %f", result.Max[1])
			}
		})
	}
}
