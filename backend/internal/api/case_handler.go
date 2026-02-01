package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sherlockos/backend/internal/db"
)

// CaseHandler handles case-related API requests
type CaseHandler struct {
	db *db.DB
}

// NewCaseHandler creates a new case handler
func NewCaseHandler(database *db.DB) *CaseHandler {
	return &CaseHandler{db: database}
}

// CreateCaseRequest represents the request body for creating a case
type CreateCaseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// Create handles POST /v1/cases
func (h *CaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	if req.Title == "" {
		BadRequest(w, "Title is required")
		return
	}

	if len(req.Title) > 200 {
		BadRequest(w, "Title must be 200 characters or less")
		return
	}

	// TODO: Implement case creation in database
	// For now, return a placeholder response
	Success(w, http.StatusCreated, map[string]interface{}{
		"id":          "case_placeholder",
		"title":       req.Title,
		"description": req.Description,
		"created_at":  "2026-02-01T00:00:00Z",
	}, nil)
}

// Get handles GET /v1/cases/{caseId}
func (h *CaseHandler) Get(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	// TODO: Implement case retrieval from database
	NotFound(w, "Case not found")
}

// GetSnapshot handles GET /v1/cases/{caseId}/snapshot
func (h *CaseHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	// TODO: Implement snapshot retrieval from database
	NotFound(w, "Snapshot not found")
}

// GetTimeline handles GET /v1/cases/{caseId}/timeline
func (h *CaseHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	// Get pagination params
	cursor := r.URL.Query().Get("cursor")
	_ = cursor // TODO: Use for pagination

	// TODO: Implement timeline retrieval from database
	Success(w, http.StatusOK, []interface{}{}, &Meta{Total: 0})
}

// UploadIntentRequest represents the request for generating presigned URLs
type UploadIntentRequest struct {
	Files []FileInfo `json:"files"`
}

// FileInfo describes a file to be uploaded
type FileInfo struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// CreateUploadIntent handles POST /v1/cases/{caseId}/upload-intent
func (h *CaseHandler) CreateUploadIntent(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	var req UploadIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	if len(req.Files) == 0 {
		BadRequest(w, "At least one file is required")
		return
	}

	// TODO: Generate presigned URLs using Supabase Storage
	Success(w, http.StatusOK, map[string]interface{}{
		"upload_batch_id": "batch_placeholder",
		"intents":         []interface{}{},
	}, nil)
}

// WitnessStatementRequest represents the request for submitting witness statements
type WitnessStatementRequest struct {
	Statements []WitnessStatement `json:"statements"`
}

// WitnessStatement represents a single witness statement
type WitnessStatement struct {
	SourceName  string  `json:"source_name"`
	Content     string  `json:"content"`
	Credibility float64 `json:"credibility"`
}

// SubmitWitnessStatements handles POST /v1/cases/{caseId}/witness-statements
func (h *CaseHandler) SubmitWitnessStatements(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	var req WitnessStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	if len(req.Statements) == 0 {
		BadRequest(w, "At least one statement is required")
		return
	}

	// Validate credibility scores
	for _, stmt := range req.Statements {
		if stmt.Credibility < 0 || stmt.Credibility > 1 {
			BadRequest(w, "Credibility must be between 0 and 1")
			return
		}
	}

	// TODO: Create commit and trigger profile job
	Success(w, http.StatusCreated, map[string]interface{}{
		"commit_id":      "commit_placeholder",
		"type":           "witness_statement",
		"profile_job_id": "job_placeholder",
	}, nil)
}

// CreateBranchRequest represents the request for creating a hypothesis branch
type CreateBranchRequest struct {
	Name         string `json:"name"`
	BaseCommitID string `json:"base_commit_id"`
}

// CreateBranch handles POST /v1/cases/{caseId}/branches
func (h *CaseHandler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	var req CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		BadRequest(w, "Branch name is required")
		return
	}

	if len(req.Name) > 100 {
		BadRequest(w, "Branch name must be 100 characters or less")
		return
	}

	// TODO: Create branch in database
	Success(w, http.StatusCreated, map[string]interface{}{
		"id":             "branch_placeholder",
		"name":           req.Name,
		"base_commit_id": req.BaseCommitID,
		"created_at":     "2026-02-01T00:00:00Z",
	}, nil)
}
