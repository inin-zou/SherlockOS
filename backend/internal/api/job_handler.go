package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
)

// JobHandler handles job-related API requests
type JobHandler struct {
	db *db.DB
}

// NewJobHandler creates a new job handler
func NewJobHandler(database *db.DB) *JobHandler {
	return &JobHandler{db: database}
}

// CreateJobRequest represents the request body for creating a job
type CreateJobRequest struct {
	Type  models.JobType         `json:"type"`
	Input map[string]interface{} `json:"input"`
}

// Create handles POST /v1/cases/{caseId}/jobs
func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	// Get idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	// Validate job type
	if !req.Type.IsValid() {
		BadRequest(w, "Invalid job type")
		return
	}

	// TODO: Check idempotency key for existing job
	_ = idempotencyKey

	// TODO: Create job in database and enqueue
	Success(w, http.StatusAccepted, map[string]interface{}{
		"job_id":     "job_placeholder",
		"type":       req.Type,
		"status":     "queued",
		"progress":   0,
		"created_at": "2026-02-01T00:00:00Z",
	}, nil)
}

// Get handles GET /v1/jobs/{jobId}
func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		BadRequest(w, "Job ID is required")
		return
	}

	// TODO: Retrieve job from database
	NotFound(w, "Job not found")
}

// CreateReasoning handles POST /v1/cases/{caseId}/reasoning
func (h *JobHandler) CreateReasoning(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	// TODO: Create reasoning job with SceneGraph input
	Success(w, http.StatusAccepted, map[string]interface{}{
		"job_id":     "job_placeholder",
		"type":       "reasoning",
		"status":     "queued",
		"progress":   0,
		"created_at": "2026-02-01T00:00:00Z",
	}, nil)
}

// CreateExport handles POST /v1/cases/{caseId}/export
func (h *JobHandler) CreateExport(w http.ResponseWriter, r *http.Request) {
	caseID := chi.URLParam(r, "caseId")
	if caseID == "" {
		BadRequest(w, "Case ID is required")
		return
	}

	// TODO: Create export job
	Success(w, http.StatusAccepted, map[string]interface{}{
		"job_id":     "job_placeholder",
		"type":       "export",
		"status":     "queued",
		"progress":   0,
		"created_at": "2026-02-01T00:00:00Z",
	}, nil)
}
