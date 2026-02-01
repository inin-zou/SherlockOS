package queue

import (
	"testing"

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
		{models.JobType("unknown"), "jobs:unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.jobType), func(t *testing.T) {
			result := getQueueName(tt.jobType)
			if result != tt.expected {
				t.Errorf("getQueueName(%v) = %v, want %v", tt.jobType, result, tt.expected)
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
}

// Note: Integration tests for Enqueue/Dequeue require a running Redis instance
// These would typically be in a separate _integration_test.go file
// or run with a build tag like `go test -tags=integration`

func TestDefaultVisibilityTimeout(t *testing.T) {
	// Verify the default visibility timeout is reasonable
	if DefaultVisibilityTimeout.Minutes() != 5 {
		t.Errorf("DefaultVisibilityTimeout = %v, want 5 minutes", DefaultVisibilityTimeout)
	}
}
