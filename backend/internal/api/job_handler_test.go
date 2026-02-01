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

func TestJobHandler_Create(t *testing.T) {
	handler := NewJobHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/jobs", handler.Create)

	tests := []struct {
		name           string
		body           interface{}
		idempotencyKey string
		wantStatus     int
	}{
		{
			name:       "invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid job type",
			body: map[string]interface{}{
				"type":  "invalid_type",
				"input": map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid reconstruction job",
			body: CreateJobRequest{
				Type: models.JobTypeReconstruction,
				Input: map[string]interface{}{
					"scan_asset_keys": []string{"key1", "key2"},
				},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "valid reasoning job",
			body: CreateJobRequest{
				Type: models.JobTypeReasoning,
				Input: map[string]interface{}{
					"scenegraph": map[string]interface{}{},
				},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "valid imagegen job",
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
			name: "valid profile job",
			body: CreateJobRequest{
				Type:  models.JobTypeProfile,
				Input: map[string]interface{}{},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "valid export job",
			body: CreateJobRequest{
				Type:  models.JobTypeExport,
				Input: map[string]interface{}{},
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "job with idempotency key",
			body: CreateJobRequest{
				Type:  models.JobTypeReconstruction,
				Input: map[string]interface{}{},
			},
			idempotencyKey: "idem_abc123",
			wantStatus:     http.StatusAccepted,
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

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/case_123/jobs", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.idempotencyKey != "" {
				req.Header.Set("Idempotency-Key", tt.idempotencyKey)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %v, want %v, body: %s", w.Code, tt.wantStatus, w.Body.String())
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
	}{
		{
			name:       "valid job ID (not found - DB not connected)",
			jobID:      "job_123",
			wantStatus: http.StatusNotFound,
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
		})
	}
}

func TestJobHandler_CreateReasoning(t *testing.T) {
	handler := NewJobHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/reasoning", handler.CreateReasoning)

	req := httptest.NewRequest(http.MethodPost, "/v1/cases/case_123/reasoning", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("CreateReasoning() status = %v, want %v", w.Code, http.StatusAccepted)
	}

	var result Response
	json.NewDecoder(w.Body).Decode(&result)

	if !result.Success {
		t.Error("CreateReasoning() should return success=true")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("CreateReasoning() data should be a map")
	}

	if data["type"] != "reasoning" {
		t.Errorf("CreateReasoning() type should be 'reasoning', got %v", data["type"])
	}
}

func TestJobHandler_CreateExport(t *testing.T) {
	handler := NewJobHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/export", handler.CreateExport)

	req := httptest.NewRequest(http.MethodPost, "/v1/cases/case_123/export", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("CreateExport() status = %v, want %v", w.Code, http.StatusAccepted)
	}

	var result Response
	json.NewDecoder(w.Body).Decode(&result)

	if !result.Success {
		t.Error("CreateExport() should return success=true")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("CreateExport() data should be a map")
	}

	if data["type"] != "export" {
		t.Errorf("CreateExport() type should be 'export', got %v", data["type"])
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
