package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCaseHandler_Create(t *testing.T) {
	handler := NewCaseHandler(nil) // DB not needed for validation tests

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "missing title",
			body:       CreateCaseRequest{Description: "Test"},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "title too long",
			body: CreateCaseRequest{
				Title: string(make([]byte, 201)),
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "valid request",
			body: CreateCaseRequest{
				Title:       "Test Case",
				Description: "Test Description",
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "valid request without description",
			body: CreateCaseRequest{
				Title: "Test Case Only Title",
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/cases", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Create(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var result map[string]interface{}
			json.NewDecoder(w.Body).Decode(&result)

			if tt.wantErr {
				if result["success"] != false {
					t.Error("Create() should return success=false on error")
				}
			} else {
				if result["success"] != true {
					t.Error("Create() should return success=true on success")
				}
			}
		})
	}
}

func TestCaseHandler_Get(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Get("/v1/cases/{caseId}", handler.Get)

	tests := []struct {
		name       string
		caseID     string
		wantStatus int
	}{
		{
			name:       "valid case ID (not found - DB not connected)",
			caseID:     "case_123",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/cases/"+tt.caseID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Get() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCaseHandler_GetTimeline(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Get("/v1/cases/{caseId}/timeline", handler.GetTimeline)

	req := httptest.NewRequest(http.MethodGet, "/v1/cases/case_123/timeline?cursor=abc&limit=10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetTimeline() status = %v, want %v", w.Code, http.StatusOK)
	}

	var result Response
	json.NewDecoder(w.Body).Decode(&result)

	if !result.Success {
		t.Error("GetTimeline() should return success=true")
	}

	if result.Meta == nil {
		t.Error("GetTimeline() should include meta")
	}
}

func TestCaseHandler_CreateUploadIntent(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/upload-intent", handler.CreateUploadIntent)

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty files array",
			body:       UploadIntentRequest{Files: []FileInfo{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid request",
			body: UploadIntentRequest{
				Files: []FileInfo{
					{Filename: "scan1.jpg", ContentType: "image/jpeg", SizeBytes: 1024000},
					{Filename: "scan2.jpg", ContentType: "image/jpeg", SizeBytes: 2048000},
				},
			},
			wantStatus: http.StatusOK,
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

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/case_123/upload-intent", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateUploadIntent() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCaseHandler_SubmitWitnessStatements(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/witness-statements", handler.SubmitWitnessStatements)

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty statements",
			body:       WitnessStatementRequest{Statements: []WitnessStatement{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credibility (below 0)",
			body: WitnessStatementRequest{
				Statements: []WitnessStatement{
					{SourceName: "Witness A", Content: "Test", Credibility: -0.5},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credibility (above 1)",
			body: WitnessStatementRequest{
				Statements: []WitnessStatement{
					{SourceName: "Witness A", Content: "Test", Credibility: 1.5},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid request",
			body: WitnessStatementRequest{
				Statements: []WitnessStatement{
					{SourceName: "Witness A", Content: "Suspect was tall, about 180cm", Credibility: 0.8},
					{SourceName: "Witness B", Content: "Saw someone running", Credibility: 0.6},
				},
			},
			wantStatus: http.StatusCreated,
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

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/case_123/witness-statements", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SubmitWitnessStatements() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCaseHandler_CreateBranch(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/branches", handler.CreateBranch)

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing name",
			body:       CreateBranchRequest{BaseCommitID: "commit_123"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "name too long",
			body: CreateBranchRequest{
				Name:         string(make([]byte, 101)),
				BaseCommitID: "commit_123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid request",
			body: CreateBranchRequest{
				Name:         "Hypothesis A",
				BaseCommitID: "commit_123",
			},
			wantStatus: http.StatusCreated,
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

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/case_123/branches", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateBranch() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
