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

func TestReplayWorker_Type(t *testing.T) {
	worker := NewReplayWorker(nil, nil, nil)
	if worker.Type() != models.JobTypeReplay {
		t.Errorf("Type() = %v, want %v", worker.Type(), models.JobTypeReplay)
	}
}

func TestReplayWorker_ProcessWithMock(t *testing.T) {
	mockClient := &clients.MockReplayClient{}
	worker := NewReplayWorker(nil, nil, mockClient)

	input := models.ReplayInput{
		CaseID:       uuid.New().String(),
		TrajectoryID: "traj_001",
		Perspective:  "third_person",
		FrameCount:   125,
		Resolution:   "480p",
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReplay,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}

func TestReplayWorker_InvalidInput(t *testing.T) {
	mockClient := &clients.MockReplayClient{}
	worker := NewReplayWorker(nil, nil, mockClient)

	// Invalid JSON input
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReplay,
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

func TestReplayWorker_MissingTrajectory(t *testing.T) {
	mockClient := &clients.MockReplayClient{}
	worker := NewReplayWorker(nil, nil, mockClient)

	// Missing required trajectory
	input := models.ReplayInput{
		CaseID:      uuid.New().String(),
		Perspective: "third_person",
		// Missing TrajectoryID and Trajectory
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReplay,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err == nil {
		t.Error("Process() expected error for missing trajectory")
	}
}

func TestReplayWorker_CustomOutput(t *testing.T) {
	customOutput := &models.ReplayOutput{
		VideoAssetKey:  "custom/video.mp4",
		ThumbnailKey:   "custom/thumb.png",
		FrameCount:     250,
		FPS:            30,
		DurationMs:     8333,
		Resolution:     "720p",
		ModelUsed:      "hy-world-1.5",
		GenerationTime: 30000,
	}

	mockClient := &clients.MockReplayClient{
		GenerateReplayFunc: func(ctx context.Context, input models.ReplayInput) (*models.ReplayOutput, error) {
			return customOutput, nil
		},
	}

	worker := NewReplayWorker(nil, nil, mockClient)

	input := models.ReplayInput{
		CaseID:       uuid.New().String(),
		TrajectoryID: "traj_001",
		Perspective:  "first_person",
		Resolution:   "720p",
	}

	inputJSON, _ := json.Marshal(input)
	job := &queue.JobMessage{
		JobID: uuid.New(),
		Type:  models.JobTypeReplay,
		Input: inputJSON,
	}

	err := worker.Process(context.Background(), job)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
}
