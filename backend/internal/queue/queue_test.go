package queue

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/models"
)

func TestGetQueueName(t *testing.T) {
	tests := []struct {
		jobType  models.JobType
		expected string
	}{
		{models.JobTypeReconstruction, QueueReconstruction},
		{models.JobTypeImageGen, QueueImageGen},
		{models.JobTypeReasoning, QueueReasoning},
		{models.JobTypeProfile, QueueProfile},
		{models.JobTypeExport, QueueExport},
		{models.JobTypeReplay, QueueReplay},
		{models.JobTypeAsset3D, QueueAsset3D},
		{models.JobTypeSceneAnalysis, QueueSceneAnalysis},
		{models.JobType("unknown"), "jobs:unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.jobType), func(t *testing.T) {
			result := GetQueueName(tt.jobType)
			if result != tt.expected {
				t.Errorf("GetQueueName(%v) = %v, want %v", tt.jobType, result, tt.expected)
			}
		})
	}
}

func TestQueueConstants(t *testing.T) {
	// Verify queue name constants follow expected pattern
	if QueueReconstruction != "jobs:reconstruction" {
		t.Errorf("QueueReconstruction = %v, want jobs:reconstruction", QueueReconstruction)
	}
	if QueueImageGen != "jobs:imagegen" {
		t.Errorf("QueueImageGen = %v, want jobs:imagegen", QueueImageGen)
	}
	if QueueReasoning != "jobs:reasoning" {
		t.Errorf("QueueReasoning = %v, want jobs:reasoning", QueueReasoning)
	}
	if QueueProfile != "jobs:profile" {
		t.Errorf("QueueProfile = %v, want jobs:profile", QueueProfile)
	}
	if QueueExport != "jobs:export" {
		t.Errorf("QueueExport = %v, want jobs:export", QueueExport)
	}
	if QueueReplay != "jobs:replay" {
		t.Errorf("QueueReplay = %v, want jobs:replay", QueueReplay)
	}
	if QueueAsset3D != "jobs:asset3d" {
		t.Errorf("QueueAsset3D = %v, want jobs:asset3d", QueueAsset3D)
	}
	if QueueSceneAnalysis != "jobs:scene_analysis" {
		t.Errorf("QueueSceneAnalysis = %v, want jobs:scene_analysis", QueueSceneAnalysis)
	}

	// Test suffixes
	if DeadLetterSuffix != ":dlq" {
		t.Errorf("DeadLetterSuffix = %v, want :dlq", DeadLetterSuffix)
	}
	if ProcessingSuffix != ":processing" {
		t.Errorf("ProcessingSuffix = %v, want :processing", ProcessingSuffix)
	}
}

func TestJobMessage_Fields(t *testing.T) {
	// Test that JobMessage struct can be created with expected fields
	msg := JobMessage{}

	// Verify zero values
	if msg.JobID.String() != "00000000-0000-0000-0000-000000000000" {
		t.Error("JobMessage JobID should be zero UUID by default")
	}

	if msg.CaseID.String() != "00000000-0000-0000-0000-000000000000" {
		t.Error("JobMessage CaseID should be zero UUID by default")
	}

	if msg.Type != "" {
		t.Error("JobMessage Type should be empty by default")
	}

	if msg.Input != nil {
		t.Error("JobMessage Input should be nil by default")
	}

	if !msg.EnqueuedAt.IsZero() {
		t.Error("JobMessage EnqueuedAt should be zero time by default")
	}

	if msg.Attempts != 0 {
		t.Error("JobMessage Attempts should be 0 by default")
	}

	if msg.LastAttempt != nil {
		t.Error("JobMessage LastAttempt should be nil by default")
	}
}

func TestJobMessage_WithValues(t *testing.T) {
	jobID := uuid.New()
	caseID := uuid.New()
	now := time.Now().UTC()

	msg := JobMessage{
		JobID:       jobID,
		CaseID:      caseID,
		Type:        models.JobTypeReconstruction,
		Input:       []byte(`{"key":"value"}`),
		EnqueuedAt:  now,
		Attempts:    2,
		LastAttempt: &now,
	}

	if msg.JobID != jobID {
		t.Errorf("JobMessage JobID = %v, want %v", msg.JobID, jobID)
	}
	if msg.CaseID != caseID {
		t.Errorf("JobMessage CaseID = %v, want %v", msg.CaseID, caseID)
	}
	if msg.Type != models.JobTypeReconstruction {
		t.Errorf("JobMessage Type = %v, want %v", msg.Type, models.JobTypeReconstruction)
	}
	if msg.Attempts != 2 {
		t.Errorf("JobMessage Attempts = %v, want 2", msg.Attempts)
	}
	if msg.LastAttempt == nil || !msg.LastAttempt.Equal(now) {
		t.Errorf("JobMessage LastAttempt = %v, want %v", msg.LastAttempt, now)
	}
}

func TestDefaultVisibilityTimeout(t *testing.T) {
	// Verify the default visibility timeout is reasonable
	if DefaultVisibilityTimeout.Minutes() != 5 {
		t.Errorf("DefaultVisibilityTimeout = %v, want 5 minutes", DefaultVisibilityTimeout)
	}
}

func TestQueueNames_DLQ(t *testing.T) {
	// Test DLQ name construction
	tests := []struct {
		jobType     models.JobType
		expectedDLQ string
	}{
		{models.JobTypeReconstruction, "jobs:reconstruction:dlq"},
		{models.JobTypeImageGen, "jobs:imagegen:dlq"},
		{models.JobTypeReasoning, "jobs:reasoning:dlq"},
		{models.JobTypeProfile, "jobs:profile:dlq"},
		{models.JobTypeExport, "jobs:export:dlq"},
	}

	for _, tt := range tests {
		t.Run(string(tt.jobType), func(t *testing.T) {
			dlqName := GetQueueName(tt.jobType) + DeadLetterSuffix
			if dlqName != tt.expectedDLQ {
				t.Errorf("DLQ name = %v, want %v", dlqName, tt.expectedDLQ)
			}
		})
	}
}

func TestQueueNames_Processing(t *testing.T) {
	// Test processing queue name construction
	tests := []struct {
		jobType            models.JobType
		expectedProcessing string
	}{
		{models.JobTypeReconstruction, "jobs:reconstruction:processing"},
		{models.JobTypeImageGen, "jobs:imagegen:processing"},
		{models.JobTypeReasoning, "jobs:reasoning:processing"},
		{models.JobTypeProfile, "jobs:profile:processing"},
		{models.JobTypeExport, "jobs:export:processing"},
	}

	for _, tt := range tests {
		t.Run(string(tt.jobType), func(t *testing.T) {
			processingName := GetQueueName(tt.jobType) + ProcessingSuffix
			if processingName != tt.expectedProcessing {
				t.Errorf("Processing queue name = %v, want %v", processingName, tt.expectedProcessing)
			}
		})
	}
}

// Note: Integration tests for Enqueue/Dequeue/Ack/Nack require a running Redis instance
// These would typically be in a separate _integration_test.go file
// or run with a build tag like `go test -tags=integration`
