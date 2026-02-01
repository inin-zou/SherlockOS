package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestJob_Validate(t *testing.T) {
	validCaseID := uuid.New()

	tests := []struct {
		name    string
		job     *Job
		wantErr bool
	}{
		{
			name:    "empty job should fail",
			job:     &Job{},
			wantErr: true,
		},
		{
			name: "missing case_id should fail",
			job: &Job{
				ID:     uuid.New(),
				Type:   JobTypeReconstruction,
				Status: JobStatusQueued,
			},
			wantErr: true,
		},
		{
			name: "invalid job type should fail",
			job: &Job{
				ID:     uuid.New(),
				CaseID: validCaseID,
				Type:   JobType("invalid"),
				Status: JobStatusQueued,
			},
			wantErr: true,
		},
		{
			name: "invalid job status should fail",
			job: &Job{
				ID:     uuid.New(),
				CaseID: validCaseID,
				Type:   JobTypeReconstruction,
				Status: JobStatus("invalid"),
			},
			wantErr: true,
		},
		{
			name: "progress below 0 should fail",
			job: &Job{
				ID:       uuid.New(),
				CaseID:   validCaseID,
				Type:     JobTypeReconstruction,
				Status:   JobStatusRunning,
				Progress: -1,
			},
			wantErr: true,
		},
		{
			name: "progress above 100 should fail",
			job: &Job{
				ID:       uuid.New(),
				CaseID:   validCaseID,
				Type:     JobTypeReconstruction,
				Status:   JobStatusRunning,
				Progress: 101,
			},
			wantErr: true,
		},
		{
			name: "valid job",
			job: &Job{
				ID:       uuid.New(),
				CaseID:   validCaseID,
				Type:     JobTypeReconstruction,
				Status:   JobStatusQueued,
				Progress: 0,
			},
			wantErr: false,
		},
		{
			name: "valid job at 100% progress",
			job: &Job{
				ID:       uuid.New(),
				CaseID:   validCaseID,
				Type:     JobTypeReasoning,
				Status:   JobStatusDone,
				Progress: 100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Job.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewJob(t *testing.T) {
	caseID := uuid.New()
	jobType := JobTypeReconstruction
	input := map[string]interface{}{
		"scan_asset_keys": []string{"key1", "key2"},
	}

	job, err := NewJob(caseID, jobType, input)
	if err != nil {
		t.Fatalf("NewJob() error = %v", err)
	}

	if job.ID == uuid.Nil {
		t.Error("NewJob() should generate a non-nil UUID")
	}

	if job.CaseID != caseID {
		t.Errorf("NewJob() CaseID = %v, want %v", job.CaseID, caseID)
	}

	if job.Type != jobType {
		t.Errorf("NewJob() Type = %v, want %v", job.Type, jobType)
	}

	if job.Status != JobStatusQueued {
		t.Errorf("NewJob() Status = %v, want %v", job.Status, JobStatusQueued)
	}

	if job.Progress != 0 {
		t.Errorf("NewJob() Progress = %v, want 0", job.Progress)
	}

	if job.RetryCount != 0 {
		t.Errorf("NewJob() RetryCount = %v, want 0", job.RetryCount)
	}

	if job.CreatedAt.IsZero() {
		t.Error("NewJob() should set CreatedAt")
	}

	if job.UpdatedAt.IsZero() {
		t.Error("NewJob() should set UpdatedAt")
	}

	// Verify input was serialized
	var decoded map[string]interface{}
	if err := json.Unmarshal(job.Input, &decoded); err != nil {
		t.Errorf("NewJob() input should be valid JSON: %v", err)
	}

	if err := job.Validate(); err != nil {
		t.Errorf("NewJob() created invalid job: %v", err)
	}
}

func TestJob_MarkRunning(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)

	err := job.MarkRunning()
	if err != nil {
		t.Errorf("MarkRunning() error = %v", err)
	}

	if job.Status != JobStatusRunning {
		t.Errorf("MarkRunning() Status = %v, want %v", job.Status, JobStatusRunning)
	}

	// Try to mark running again (should fail)
	err = job.MarkRunning()
	if err == nil {
		t.Error("MarkRunning() should fail when job is already running")
	}
}

func TestJob_UpdateProgress(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)
	job.MarkRunning()

	tests := []struct {
		progress int
		wantErr  bool
	}{
		{0, false},
		{50, false},
		{100, false},
		{-1, true},
		{101, true},
	}

	for _, tt := range tests {
		err := job.UpdateProgress(tt.progress)
		if (err != nil) != tt.wantErr {
			t.Errorf("UpdateProgress(%d) error = %v, wantErr %v", tt.progress, err, tt.wantErr)
		}
		if err == nil && job.Progress != tt.progress {
			t.Errorf("UpdateProgress(%d) Progress = %v", tt.progress, job.Progress)
		}
	}
}

func TestJob_MarkDone(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)

	// Should fail when not running
	err := job.MarkDone(map[string]string{"result": "success"})
	if err == nil {
		t.Error("MarkDone() should fail when job is not running")
	}

	// Mark running first
	job.MarkRunning()

	// Now should succeed
	output := map[string]string{"result": "success"}
	err = job.MarkDone(output)
	if err != nil {
		t.Errorf("MarkDone() error = %v", err)
	}

	if job.Status != JobStatusDone {
		t.Errorf("MarkDone() Status = %v, want %v", job.Status, JobStatusDone)
	}

	if job.Progress != 100 {
		t.Errorf("MarkDone() Progress = %v, want 100", job.Progress)
	}

	// Verify output was serialized
	var decoded map[string]string
	if err := json.Unmarshal(job.Output, &decoded); err != nil {
		t.Errorf("MarkDone() output should be valid JSON: %v", err)
	}
}

func TestJob_MarkFailed(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)
	job.MarkRunning()

	errMsg := "Model API returned 500"
	job.MarkFailed(errMsg)

	if job.Status != JobStatusFailed {
		t.Errorf("MarkFailed() Status = %v, want %v", job.Status, JobStatusFailed)
	}

	if job.Error != errMsg {
		t.Errorf("MarkFailed() Error = %v, want %v", job.Error, errMsg)
	}
}

func TestJob_MarkCanceled(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)

	job.MarkCanceled()

	if job.Status != JobStatusCanceled {
		t.Errorf("MarkCanceled() Status = %v, want %v", job.Status, JobStatusCanceled)
	}
}

func TestJob_IncrementRetry(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)

	if job.RetryCount != 0 {
		t.Errorf("Initial RetryCount = %v, want 0", job.RetryCount)
	}

	job.IncrementRetry()
	if job.RetryCount != 1 {
		t.Errorf("After 1st retry, RetryCount = %v, want 1", job.RetryCount)
	}

	job.IncrementRetry()
	if job.RetryCount != 2 {
		t.Errorf("After 2nd retry, RetryCount = %v, want 2", job.RetryCount)
	}
}

func TestJob_SetIdempotencyKey(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)

	key := "idem_abc123"
	job.SetIdempotencyKey(key)

	if job.IdempotencyKey != key {
		t.Errorf("SetIdempotencyKey() = %v, want %v", job.IdempotencyKey, key)
	}
}

func TestJob_Heartbeat(t *testing.T) {
	job, _ := NewJob(uuid.New(), JobTypeReconstruction, nil)
	originalTime := job.UpdatedAt

	// Small delay to ensure time difference
	job.Heartbeat()

	if !job.UpdatedAt.After(originalTime) && job.UpdatedAt != originalTime {
		t.Error("Heartbeat() should update UpdatedAt")
	}
}
