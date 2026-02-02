package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/queue"
)

// RegisterRoutes sets up all API routes
func RegisterRoutes(r chi.Router, database *db.DB) {
	RegisterRoutesWithQueue(r, database, nil)
}

// RegisterRoutesWithQueue sets up all API routes with queue support
func RegisterRoutesWithQueue(r chi.Router, database *db.DB, q queue.JobQueue) {
	// Initialize handlers
	caseHandler := NewCaseHandlerWithQueue(database, q)
	var jobHandler *JobHandler
	if q != nil {
		jobHandler = NewJobHandlerWithQueue(database, q)
	} else {
		jobHandler = NewJobHandler(database)
	}

	// Cases
	r.Route("/cases", func(r chi.Router) {
		r.Post("/", caseHandler.Create)
		r.Get("/{caseId}", caseHandler.Get)
		r.Get("/{caseId}/snapshot", caseHandler.GetSnapshot)
		r.Get("/{caseId}/timeline", caseHandler.GetTimeline)
		r.Post("/{caseId}/upload-intent", caseHandler.CreateUploadIntent)
		r.Post("/{caseId}/jobs", jobHandler.Create)
		r.Post("/{caseId}/witness-statements", caseHandler.SubmitWitnessStatements)
		r.Post("/{caseId}/branches", caseHandler.CreateBranch)
		r.Post("/{caseId}/reasoning", jobHandler.CreateReasoning)
		r.Post("/{caseId}/export", jobHandler.CreateExport)
	})

	// Jobs
	r.Route("/jobs", func(r chi.Router) {
		r.Get("/{jobId}", jobHandler.Get)
	})
}
