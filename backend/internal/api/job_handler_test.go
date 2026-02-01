package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sherlockos/backend/internal/models"
)

const testJobID = "770e8400-e29b-41d4-a716-446655440000"

func TestJobHandler_Create(t *testing.T) {
	handler := NewJobHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/jobs", handler.Create)

	tests := []struct {
		name           string
		caseID         string
		body           interface{}
		idempotencyKey string
		wantStatus     int
		wantErr        string
	}{
		{
			name:       "invalid JSON",
			caseID:     testCaseID,
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid request body",
		},
		{
			name:   "invalid job type",
			caseID: testCaseID,
			body: map[string]interface{}{
				"type":  "invalid_type",
				"input": map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid job type",
		},
		{
			name:   "valid reconstruction job",
			caseID: testCaseID,
			body: CreateJobRequest{
				Type: models.JobTypeReconstruction,
				Input: map[string]interface{}{
					"scan_asset_keys": []string{"key1", "key2"},
				},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:   "valid reasoning job",
			caseID: testCaseID,
			body: CreateJobRequest{
				Type: models.JobTypeReasoning,
				Input: map[string]interface{}{
					"scenegraph": map[string]interface{}{},
				},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:   "valid imagegen job",
			caseID: testCaseID,
			body: CreateJobRequest{
				Type: models.JobTypeImageGen,
				Input: map[string]interface{}{
					"gen_type":   "portrait",
					"resolution": "1k",
				},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:   "valid profile job",
			caseID: testCaseID,
			body: CreateJobRequest{
				Type:  models.JobTypeProfile,
				Input: map[string]interface{}{},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:   "valid export job",
			caseID: testCaseID,
			body: CreateJobRequest{
				Type:  models.JobTypeExport,
				Input: map[string]interface{}{},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:   "job with idempotency key",
			caseID: testCaseID,
			body: CreateJobRequest{
				Type:  models.JobTypeReconstruction,
				Input: map[string]interface{}{},
			},
			idempotencyKey: "idem_abc123",
			wantStatus:     http.StatusAccepted,
		},
		{
			name:   "invalid case ID format",
			caseID: "invalid-case-id",
			body: CreateJobRequest{
				Type:  models.JobTypeReconstruction,
				Input: map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid case ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/"+tt.caseID+"/jobs", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.idempotencyKey != "" {
				req.Header.Set("Idempotency-Key", tt.idempotencyKey)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %v, want %v, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantErr != "" {
				errMsg := getErrorMessage(w.Body.Bytes())
				if errMsg != tt.wantErr {
					t.Errorf("Create() error = %v, want %v", errMsg, tt.wantErr)
				}
			}

			if tt.wantStatus == http.StatusAccepted {
				var result Response
				json.NewDecoder(w.Body).Decode(&result)

				if !result.Success {
					t.Error("Create() should return success=true")
				}

				data, ok := result.Data.(map[string]interface{})
				if !ok {
					t.Fatal("Create() data should be a map")
				}

				if data["job_id"] == "" {
					t.Error("Create() should return job_id")
				}

				if data["status"] != "queued" {
					t.Errorf("Create() status should be 'queued', got %v", data["status"])
				}

				if data["progress"] != float64(0) {
					t.Errorf("Create() progress should be 0, got %v", data["progress"])
				}
			}
		})
	}
}

func TestJobHandler_Get(t *testing.T) {
	handler := NewJobHandler(nil)

	r := chi.NewRouter()
	r.Get("/v1/jobs/{jobId}", handler.Get)

	tests := []struct {
		name       string
		jobID      string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "valid UUID but not found (DB not connected)",
			jobID:      testJobID,
			wantStatus: http.StatusNotFound,
			wantErr:    "Job not found",
		},
		{
			name:       "invalid UUID format",
			jobID:      "job_123",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid job ID format",
		},
		{
			name:       "random string",
			jobID:      "not-a-uuid",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid job ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/jobs/"+tt.jobID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Get() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantErr != "" {
				errMsg := getErrorMessage(w.Body.Bytes())
				if errMsg != tt.wantErr {
					t.Errorf("Get() error = %v, want %v", errMsg, tt.wantErr)
				}
			}
		})
	}
}

func TestJobHandler_CreateReasoning(t *testing.T) {
	handler := NewJobHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/reasoning", handler.CreateReasoning)

	tests := []struct {
		name       string
		caseID     string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "valid request",
			caseID:     testCaseID,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid case ID",
			caseID:     "invalid",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid case ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/cases/"+tt.caseID+"/reasoning", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateReasoning() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusAccepted {
				var result Response
				json.NewDecoder(w.Body).Decode(&result)

				if !result.Success {
					t.Error("CreateReasoning() should return success=true")
				}

				data, ok := result.Data.(map[string]interface{})
				if !ok {
					t.Fatal("CreateReasoning() data should be a map")
				}

				if data["job_id"] == "" {
					t.Error("CreateReasoning() should return job_id")
				}

				if data["type"] != "reasoning" {
					t.Errorf("CreateReasoning() type should be 'reasoning', got %v", data["type"])
				}
			}
		})
	}
}

func TestJobHandler_CreateExport(t *testing.T) {
	handler := NewJobHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/export", handler.CreateExport)

	tests := []struct {
		name       string
		caseID     string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "valid request",
			caseID:     testCaseID,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid case ID",
			caseID:     "invalid",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid case ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/cases/"+tt.caseID+"/export", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateExport() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusAccepted {
				var result Response
				json.NewDecoder(w.Body).Decode(&result)

				if !result.Success {
					t.Error("CreateExport() should return success=true")
				}

				data, ok := result.Data.(map[string]interface{})
				if !ok {
					t.Fatal("CreateExport() data should be a map")
				}

				if data["job_id"] == "" {
					t.Error("CreateExport() should return job_id")
				}

				if data["type"] != "export" {
					t.Errorf("CreateExport() type should be 'export', got %v", data["type"])
				}
			}
		})
	}
}

func TestJobType_Validation(t *testing.T) {
	validTypes := []models.JobType{
		models.JobTypeReconstruction,
		models.JobTypeImageGen,
		models.JobTypeReasoning,
		models.JobTypeProfile,
		models.JobTypeExport,
	}

	for _, jt := range validTypes {
		if !jt.IsValid() {
			t.Errorf("JobType %v should be valid", jt)
		}
	}

	invalidType := models.JobType("invalid")
	if invalidType.IsValid() {
		t.Error("Invalid JobType should not be valid")
	}
}
