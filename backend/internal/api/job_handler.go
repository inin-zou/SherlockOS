package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
	"github.com/sherlockos/backend/internal/workers"
)

// JobHandler handles job-related API requests
type JobHandler struct {
	repo  *db.Repository
	queue queue.JobQueue
}

// NewJobHandler creates a new job handler
func NewJobHandler(database *db.DB) *JobHandler {
	var repo *db.Repository
	if database != nil {
		repo = db.NewRepository(database)
	}
	return &JobHandler{repo: repo, queue: nil}
}

// NewJobHandlerWithQueue creates a new job handler with queue support
func NewJobHandlerWithQueue(database *db.DB, q queue.JobQueue) *JobHandler {
	var repo *db.Repository
	if database != nil {
		repo = db.NewRepository(database)
	}
	return &JobHandler{repo: repo, queue: q}
}

// CreateJobRequest represents the request body for creating a job
type CreateJobRequest struct {
	Type  models.JobType         `json:"type"`
	Input map[string]interface{} `json:"input"`
}

// Create handles POST /v1/cases/{caseId}/jobs
func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	// Check if a worker is available for this job type
	registry := workers.GetGlobalRegistry()
	if !registry.IsAvailable(req.Type) {
		reason := workers.GetUnavailableReason(req.Type)
		ServiceUnavailable(w, fmt.Sprintf("Service not available for job type '%s': %s", req.Type, reason))
		return
	}

	// Check idempotency key for existing job
	if h.repo != nil && idempotencyKey != "" {
		existingJob, _ := h.repo.GetJobByIdempotencyKey(r.Context(), idempotencyKey)
		if existingJob != nil {
			// Return existing job
			Success(w, http.StatusOK, map[string]interface{}{
				"job_id":     existingJob.ID.String(),
				"type":       existingJob.Type,
				"status":     existingJob.Status,
				"progress":   existingJob.Progress,
				"created_at": existingJob.CreatedAt.Format(time.RFC3339),
			}, nil)
			return
		}
	}

	// Ensure case_id is in the input for job types that need it
	if req.Input == nil {
		req.Input = make(map[string]interface{})
	}
	if _, ok := req.Input["case_id"]; !ok {
		req.Input["case_id"] = caseID.String()
	}

	// Create job
	job, err := models.NewJob(caseID, req.Type, req.Input)
	if err != nil {
		InternalError(w, "Failed to create job")
		return
	}
	if idempotencyKey != "" {
		job.SetIdempotencyKey(idempotencyKey)
	}

	// Save to database if repo is available
	if h.repo != nil {
		if err := h.repo.CreateJob(r.Context(), job); err != nil {
			InternalError(w, "Failed to save job")
			return
		}
	}

	// Enqueue job for processing
	if h.queue != nil {
		if err := h.queue.Enqueue(r.Context(), job); err != nil {
			// Log error but don't fail - job is saved in DB
			// Workers can pick it up via zombie recovery
		}
	}

	Success(w, http.StatusAccepted, map[string]interface{}{
		"job_id":     job.ID.String(),
		"type":       job.Type,
		"status":     job.Status,
		"progress":   job.Progress,
		"created_at": job.CreatedAt.Format(time.RFC3339),
	}, nil)
}

// Get handles GET /v1/jobs/{jobId}
func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request) {
	jobIDStr := chi.URLParam(r, "jobId")
	if jobIDStr == "" {
		BadRequest(w, "Job ID is required")
		return
	}

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		BadRequest(w, "Invalid job ID format")
		return
	}

	if h.repo == nil {
		NotFound(w, "Job not found")
		return
	}

	job, err := h.repo.GetJob(r.Context(), jobID)
	if err != nil {
		log.Printf("Failed to retrieve job %s: %v", jobID, err)
		InternalError(w, "Failed to retrieve job")
		return
	}
	if job == nil {
		NotFound(w, "Job not found")
		return
	}

	response := map[string]interface{}{
		"job_id":     job.ID.String(),
		"case_id":    job.CaseID.String(),
		"type":       job.Type,
		"status":     job.Status,
		"progress":   job.Progress,
		"created_at": job.CreatedAt.Format(time.RFC3339),
		"updated_at": job.UpdatedAt.Format(time.RFC3339),
	}

	if len(job.Output) > 0 {
		response["output"] = json.RawMessage(job.Output)
	}
	if job.Error != "" {
		response["error"] = job.Error
	}

	Success(w, http.StatusOK, response, nil)
}

// CreateReasoning handles POST /v1/cases/{caseId}/reasoning
func (h *JobHandler) CreateReasoning(w http.ResponseWriter, r *http.Request) {
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

	// Get current scenegraph to include in job input
	var scenegraph *models.SceneGraph
	if h.repo != nil {
		snapshot, _ := h.repo.GetSceneSnapshot(r.Context(), caseID)
		if snapshot != nil {
			scenegraph = snapshot.Scenegraph
		}
	}
	if scenegraph == nil {
		scenegraph = models.NewEmptySceneGraph()
	}

	// Create reasoning job with SceneGraph input
	job, err := models.NewJob(caseID, models.JobTypeReasoning, map[string]interface{}{
		"case_id":    caseID.String(),
		"scenegraph": scenegraph,
	})
	if err != nil {
		InternalError(w, "Failed to create reasoning job")
		return
	}

	if h.repo != nil {
		if err := h.repo.CreateJob(r.Context(), job); err != nil {
			InternalError(w, "Failed to save reasoning job")
			return
		}
	}

	// Enqueue job for processing
	if h.queue != nil {
		h.queue.Enqueue(r.Context(), job)
	}

	Success(w, http.StatusAccepted, map[string]interface{}{
		"job_id":     job.ID.String(),
		"type":       job.Type,
		"status":     job.Status,
		"progress":   job.Progress,
		"created_at": job.CreatedAt.Format(time.RFC3339),
	}, nil)
}

// CreateExport handles POST /v1/cases/{caseId}/export
func (h *JobHandler) CreateExport(w http.ResponseWriter, r *http.Request) {
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

	// Create export job
	job, err := models.NewJob(caseID, models.JobTypeExport, map[string]interface{}{
		"format": "pdf",
	})
	if err != nil {
		InternalError(w, "Failed to create export job")
		return
	}

	if h.repo != nil {
		if err := h.repo.CreateJob(r.Context(), job); err != nil {
			log.Printf("Failed to save export job: %v", err)
			InternalError(w, "Failed to save export job")
			return
		}
	}

	// Enqueue job for processing
	if h.queue != nil {
		h.queue.Enqueue(r.Context(), job)
	}

	Success(w, http.StatusAccepted, map[string]interface{}{
		"job_id":     job.ID.String(),
		"type":       job.Type,
		"status":     job.Status,
		"progress":   job.Progress,
		"created_at": job.CreatedAt.Format(time.RFC3339),
	}, nil)
}
