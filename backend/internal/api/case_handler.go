package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

// CaseHandler handles case-related API requests
type CaseHandler struct {
	repo  *db.Repository
	queue queue.JobQueue
}

// NewCaseHandler creates a new case handler
func NewCaseHandler(database *db.DB) *CaseHandler {
	var repo *db.Repository
	if database != nil {
		repo = db.NewRepository(database)
	}
	return &CaseHandler{repo: repo, queue: nil}
}

// NewCaseHandlerWithQueue creates a new case handler with queue support
func NewCaseHandlerWithQueue(database *db.DB, q queue.JobQueue) *CaseHandler {
	var repo *db.Repository
	if database != nil {
		repo = db.NewRepository(database)
	}
	return &CaseHandler{repo: repo, queue: q}
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

	// Create case model
	c := models.NewCase(req.Title, req.Description)

	// Save to database if repo is available
	if h.repo != nil {
		if err := h.repo.CreateCase(r.Context(), c); err != nil {
			InternalError(w, "Failed to create case")
			return
		}

		// Create initial empty scene snapshot
		snapshot := models.NewSceneSnapshot(c.ID, uuid.Nil, models.NewEmptySceneGraph())
		// Note: commit_id is nil for initial snapshot, will be updated on first commit
	_ = snapshot // Will be created with first commit
	}

	Success(w, http.StatusCreated, map[string]interface{}{
		"id":          c.ID.String(),
		"title":       c.Title,
		"description": c.Description,
		"created_at":  c.CreatedAt.Format(time.RFC3339),
	}, nil)
}

// List handles GET /v1/cases
func (h *CaseHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		Success(w, http.StatusOK, []interface{}{}, nil)
		return
	}

	cases, err := h.repo.ListCases(r.Context(), 100, nil)
	if err != nil {
		InternalError(w, "Failed to list cases")
		return
	}

	// Convert to response format
	result := make([]map[string]interface{}, 0, len(cases))
	for _, c := range cases {
		result = append(result, map[string]interface{}{
			"id":          c.ID.String(),
			"title":       c.Title,
			"description": c.Description,
			"created_at":  c.CreatedAt.Format(time.RFC3339),
		})
	}

	Success(w, http.StatusOK, result, nil)
}

// Get handles GET /v1/cases/{caseId}
func (h *CaseHandler) Get(w http.ResponseWriter, r *http.Request) {
	caseIDStr := chi.URLParam(r, "caseId")
	if caseIDStr == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		BadRequest(w, "Invalid case ID format")
		return
	}

	if h.repo == nil {
		NotFound(w, "Case not found")
		return
	}

	c, err := h.repo.GetCase(r.Context(), caseID)
	if err != nil {
		InternalError(w, "Failed to retrieve case")
		return
	}
	if c == nil {
		NotFound(w, "Case not found")
		return
	}

	Success(w, http.StatusOK, map[string]interface{}{
		"id":          c.ID.String(),
		"title":       c.Title,
		"description": c.Description,
		"created_at":  c.CreatedAt.Format(time.RFC3339),
	}, nil)
}

// GetSnapshot handles GET /v1/cases/{caseId}/snapshot
func (h *CaseHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	caseIDStr := chi.URLParam(r, "caseId")
	if caseIDStr == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		BadRequest(w, "Invalid case ID format")
		return
	}

	if h.repo == nil {
		NotFound(w, "Snapshot not found")
		return
	}

	snapshot, err := h.repo.GetSceneSnapshot(r.Context(), caseID)
	if err != nil {
		InternalError(w, "Failed to retrieve snapshot")
		return
	}
	if snapshot == nil {
		// Return empty scenegraph if no snapshot exists yet
		Success(w, http.StatusOK, map[string]interface{}{
			"case_id":    caseID.String(),
			"commit_id":  nil,
			"scenegraph": models.NewEmptySceneGraph(),
			"updated_at": time.Now().Format(time.RFC3339),
		}, nil)
		return
	}

	Success(w, http.StatusOK, map[string]interface{}{
		"case_id":    snapshot.CaseID.String(),
		"commit_id":  snapshot.CommitID.String(),
		"scenegraph": snapshot.Scenegraph,
		"updated_at": snapshot.UpdatedAt.Format(time.RFC3339),
	}, nil)
}

// GetTimeline handles GET /v1/cases/{caseId}/timeline
func (h *CaseHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	caseIDStr := chi.URLParam(r, "caseId")
	if caseIDStr == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		BadRequest(w, "Invalid case ID format")
		return
	}

	// Parse pagination params
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	var cursor *time.Time
	if cursorStr != "" {
		if t, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			cursor = &t
		}
	}

	if h.repo == nil {
		Success(w, http.StatusOK, []interface{}{}, &Meta{Total: 0})
		return
	}

	commits, err := h.repo.GetCommitsByCase(r.Context(), caseID, limit, cursor)
	if err != nil {
		InternalError(w, "Failed to retrieve timeline")
		return
	}

	// Convert to response format
	var result []map[string]interface{}
	for _, c := range commits {
		item := map[string]interface{}{
			"id":         c.ID.String(),
			"case_id":    c.CaseID.String(),
			"type":       c.Type,
			"summary":    c.Summary,
			"payload":    json.RawMessage(c.Payload),
			"created_at": c.CreatedAt.Format(time.RFC3339),
		}
		if c.ParentCommitID != nil {
			item["parent_commit_id"] = c.ParentCommitID.String()
		}
		if c.BranchID != nil {
			item["branch_id"] = c.BranchID.String()
		}
		result = append(result, item)
	}

	// Set next cursor
	var nextCursor string
	if len(commits) == limit {
		nextCursor = commits[len(commits)-1].CreatedAt.Format(time.RFC3339)
	}

	Success(w, http.StatusOK, result, &Meta{Cursor: nextCursor})
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
	caseIDStr := chi.URLParam(r, "caseId")
	if caseIDStr == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		BadRequest(w, "Invalid case ID format")
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

	// Generate batch ID
	batchID := uuid.New().String()

	// Generate upload intents (placeholder - would use Supabase Storage SDK)
	var intents []map[string]interface{}
	for _, f := range req.Files {
		storageKey := "cases/" + caseID.String() + "/scans/" + batchID + "/" + f.Filename
		intents = append(intents, map[string]interface{}{
			"filename":      f.Filename,
			"storage_key":   storageKey,
			"presigned_url": "https://placeholder.supabase.co/storage/v1/upload/" + storageKey,
			"expires_at":    time.Now().Add(30 * time.Minute).Format(time.RFC3339),
		})
	}

	Success(w, http.StatusOK, map[string]interface{}{
		"upload_batch_id": batchID,
		"intents":         intents,
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
	caseIDStr := chi.URLParam(r, "caseId")
	if caseIDStr == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		BadRequest(w, "Invalid case ID format")
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

	// Create commit for witness statements
	payload := map[string]interface{}{
		"statements": req.Statements,
	}

	commit, err := models.NewCommit(caseID, models.CommitTypeWitnessStatement, "Added witness statements", payload)
	if err != nil {
		InternalError(w, "Failed to create commit")
		return
	}

	// Get parent commit
	if h.repo != nil {
		latestCommit, _ := h.repo.GetLatestCommit(r.Context(), caseID)
		if latestCommit != nil {
			commit.SetParent(latestCommit.ID)
		}

		if err := h.repo.CreateCommit(r.Context(), commit); err != nil {
			InternalError(w, "Failed to save commit")
			return
		}

		// Create profile job
		profileJob, _ := models.NewJob(caseID, models.JobTypeProfile, map[string]interface{}{
			"case_id":    caseID.String(),
			"statements": req.Statements,
			"commit_id":  commit.ID.String(),
		})
		h.repo.CreateJob(r.Context(), profileJob)

		// Enqueue profile job for processing
		if h.queue != nil {
			h.queue.Enqueue(r.Context(), profileJob)
		}

		Success(w, http.StatusCreated, map[string]interface{}{
			"commit_id":      commit.ID.String(),
			"type":           "witness_statement",
			"profile_job_id": profileJob.ID.String(),
		}, nil)
		return
	}

	Success(w, http.StatusCreated, map[string]interface{}{
		"commit_id":      commit.ID.String(),
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
	caseIDStr := chi.URLParam(r, "caseId")
	if caseIDStr == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		BadRequest(w, "Invalid case ID format")
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

	// Parse base commit ID
	var baseCommitID uuid.UUID
	if req.BaseCommitID != "" {
		baseCommitID, err = uuid.Parse(req.BaseCommitID)
		if err != nil {
			BadRequest(w, "Invalid base commit ID format")
			return
		}
	} else if h.repo != nil {
		// Use latest commit as base
		latestCommit, _ := h.repo.GetLatestCommit(r.Context(), caseID)
		if latestCommit != nil {
			baseCommitID = latestCommit.ID
		}
	}

	// Create branch
	branch := models.NewBranch(caseID, req.Name, baseCommitID)

	if h.repo != nil {
		if err := h.repo.CreateBranch(r.Context(), branch); err != nil {
			InternalError(w, "Failed to create branch")
			return
		}
	}

	Success(w, http.StatusCreated, map[string]interface{}{
		"id":             branch.ID.String(),
		"name":           branch.Name,
		"base_commit_id": branch.BaseCommitID.String(),
		"created_at":     branch.CreatedAt.Format(time.RFC3339),
	}, nil)
}
