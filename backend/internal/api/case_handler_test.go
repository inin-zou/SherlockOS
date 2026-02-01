package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

const testCaseID = "550e8400-e29b-41d4-a716-446655440000"
const testCommitID = "660e8400-e29b-41d4-a716-446655440000"

// helper to extract error message from response
func getErrorMessage(body []byte) string {
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		if msg, ok := errObj["message"].(string); ok {
			return msg
		}
	}
	return ""
}

// helper to extract data from response
func getData(body []byte) map[string]interface{} {
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)
	if data, ok := resp["data"].(map[string]interface{}); ok {
		return data
	}
	return nil
}

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
				Title: strings.Repeat("a", 201),
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
		wantErr    string
	}{
		{
			name:       "valid UUID but not found (DB not connected)",
			caseID:     testCaseID,
			wantStatus: http.StatusNotFound,
			wantErr:    "Case not found",
		},
		{
			name:       "invalid UUID format",
			caseID:     "case_123",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid case ID format",
		},
		{
			name:       "random string",
			caseID:     "not-a-uuid-at-all",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid case ID format",
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

			if tt.wantErr != "" {
				errMsg := getErrorMessage(w.Body.Bytes())
				if errMsg != tt.wantErr {
					t.Errorf("Get() error = %v, want %v", errMsg, tt.wantErr)
				}
			}
		})
	}
}

func TestCaseHandler_GetSnapshot(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Get("/v1/cases/{caseId}/snapshot", handler.GetSnapshot)

	tests := []struct {
		name       string
		caseID     string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "valid UUID but no DB",
			caseID:     testCaseID,
			wantStatus: http.StatusNotFound,
			wantErr:    "Snapshot not found",
		},
		{
			name:       "invalid UUID",
			caseID:     "invalid",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid case ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/cases/"+tt.caseID+"/snapshot", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetSnapshot() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCaseHandler_GetTimeline(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Get("/v1/cases/{caseId}/timeline", handler.GetTimeline)

	tests := []struct {
		name       string
		caseID     string
		query      string
		wantStatus int
	}{
		{
			name:       "valid request returns empty (no DB)",
			caseID:     testCaseID,
			query:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "with valid limit",
			caseID:     testCaseID,
			query:      "?limit=25",
			wantStatus: http.StatusOK,
		},
		{
			name:       "with cursor",
			caseID:     testCaseID,
			query:      "?cursor=2026-01-01T00:00:00Z",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID",
			caseID:     "invalid",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/cases/"+tt.caseID+"/timeline"+tt.query, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetTimeline() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var result Response
				json.NewDecoder(w.Body).Decode(&result)

				if !result.Success {
					t.Error("GetTimeline() should return success=true")
				}

				if result.Meta == nil {
					t.Error("GetTimeline() should include meta")
				}
			}
		})
	}
}

func TestCaseHandler_CreateUploadIntent(t *testing.T) {
	handler := NewCaseHandler(nil)

	r := chi.NewRouter()
	r.Post("/v1/cases/{caseId}/upload-intent", handler.CreateUploadIntent)

	tests := []struct {
		name       string
		caseID     string
		body       interface{}
		wantStatus int
		wantErr    string
	}{
		{
			name:       "invalid JSON",
			caseID:     testCaseID,
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid request body",
		},
		{
			name:       "empty files array",
			caseID:     testCaseID,
			body:       UploadIntentRequest{Files: []FileInfo{}},
			wantStatus: http.StatusBadRequest,
			wantErr:    "At least one file is required",
		},
		{
			name:   "valid request",
			caseID: testCaseID,
			body: UploadIntentRequest{
				Files: []FileInfo{
					{Filename: "scan1.jpg", ContentType: "image/jpeg", SizeBytes: 1024000},
					{Filename: "scan2.jpg", ContentType: "image/jpeg", SizeBytes: 2048000},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid case ID",
			caseID:     "invalid",
			body:       UploadIntentRequest{Files: []FileInfo{{Filename: "test.jpg"}}},
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

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/"+tt.caseID+"/upload-intent", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateUploadIntent() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				data := getData(w.Body.Bytes())
				if data == nil || data["upload_batch_id"] == nil {
					t.Error("CreateUploadIntent() should return upload_batch_id")
				}
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
		caseID     string
		body       interface{}
		wantStatus int
		wantErr    string
	}{
		{
			name:       "invalid JSON",
			caseID:     testCaseID,
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid request body",
		},
		{
			name:       "empty statements",
			caseID:     testCaseID,
			body:       WitnessStatementRequest{Statements: []WitnessStatement{}},
			wantStatus: http.StatusBadRequest,
			wantErr:    "At least one statement is required",
		},
		{
			name:   "invalid credibility (below 0)",
			caseID: testCaseID,
			body: WitnessStatementRequest{
				Statements: []WitnessStatement{
					{SourceName: "Witness A", Content: "Test", Credibility: -0.5},
				},
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Credibility must be between 0 and 1",
		},
		{
			name:   "invalid credibility (above 1)",
			caseID: testCaseID,
			body: WitnessStatementRequest{
				Statements: []WitnessStatement{
					{SourceName: "Witness A", Content: "Test", Credibility: 1.5},
				},
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Credibility must be between 0 and 1",
		},
		{
			name:   "valid request",
			caseID: testCaseID,
			body: WitnessStatementRequest{
				Statements: []WitnessStatement{
					{SourceName: "Witness A", Content: "Suspect was tall, about 180cm", Credibility: 0.8},
					{SourceName: "Witness B", Content: "Saw someone running", Credibility: 0.6},
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "invalid case ID",
			caseID: "invalid",
			body: WitnessStatementRequest{
				Statements: []WitnessStatement{
					{SourceName: "Witness", Content: "Test", Credibility: 0.5},
				},
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

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/"+tt.caseID+"/witness-statements", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SubmitWitnessStatements() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusCreated {
				data := getData(w.Body.Bytes())
				if data == nil || data["commit_id"] == nil {
					t.Error("SubmitWitnessStatements() should return commit_id")
				}
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
		caseID     string
		body       interface{}
		wantStatus int
		wantErr    string
	}{
		{
			name:       "invalid JSON",
			caseID:     testCaseID,
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid request body",
		},
		{
			name:       "missing name",
			caseID:     testCaseID,
			body:       CreateBranchRequest{BaseCommitID: testCommitID},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Branch name is required",
		},
		{
			name:   "name too long",
			caseID: testCaseID,
			body: CreateBranchRequest{
				Name:         strings.Repeat("a", 101),
				BaseCommitID: testCommitID,
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Branch name must be 100 characters or less",
		},
		{
			name:   "valid request",
			caseID: testCaseID,
			body: CreateBranchRequest{
				Name:         "Hypothesis A",
				BaseCommitID: testCommitID,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "valid request without base commit",
			caseID: testCaseID,
			body: CreateBranchRequest{
				Name: "New Branch",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "invalid case ID",
			caseID: "invalid",
			body: CreateBranchRequest{
				Name: "Test",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid case ID format",
		},
		{
			name:   "invalid base commit ID",
			caseID: testCaseID,
			body: CreateBranchRequest{
				Name:         "Test",
				BaseCommitID: "not-a-uuid",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid base commit ID format",
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

			req := httptest.NewRequest(http.MethodPost, "/v1/cases/"+tt.caseID+"/branches", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateBranch() status = %v, want %v, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantStatus == http.StatusCreated {
				data := getData(w.Body.Bytes())
				if data == nil || data["id"] == nil {
					t.Error("CreateBranch() should return id")
				}
				if data == nil || data["name"] == nil {
					t.Error("CreateBranch() should return name")
				}
			}
		})
	}
}
