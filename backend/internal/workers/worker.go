package workers

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

// Worker interface defines the contract for all workers
type Worker interface {
	// Type returns the job type this worker handles
	Type() models.JobType

	// Process handles a single job
	Process(ctx context.Context, job *queue.JobMessage) error
}

// Manager manages worker lifecycle
type Manager struct {
	db       *db.DB
	queue    *queue.Queue
	workers  map[models.JobType]Worker
	wg       sync.WaitGroup
	shutdown chan struct{}
}

// NewManager creates a new worker manager
func NewManager(database *db.DB, q *queue.Queue) *Manager {
	return &Manager{
		db:       database,
		queue:    q,
		workers:  make(map[models.JobType]Worker),
		shutdown: make(chan struct{}),
	}
}

// Register adds a worker for a specific job type
func (m *Manager) Register(w Worker) {
	m.workers[w.Type()] = w
}

// Start begins processing jobs for all registered workers
func (m *Manager) Start(ctx context.Context) {
	for jobType, worker := range m.workers {
		m.wg.Add(1)
		go m.runWorker(ctx, jobType, worker)
	}
}

// Stop gracefully shuts down all workers
func (m *Manager) Stop() {
	close(m.shutdown)
	m.wg.Wait()
}

// runWorker runs a single worker in a loop
func (m *Manager) runWorker(ctx context.Context, jobType models.JobType, w Worker) {
	defer m.wg.Done()

	log.Printf("Starting worker for %s jobs", jobType)

	for {
		select {
		case <-m.shutdown:
			log.Printf("Shutting down worker for %s jobs", jobType)
			return
		case <-ctx.Done():
			return
		default:
			// Try to dequeue a job with 5 second timeout
			job, err := m.queue.Dequeue(ctx, jobType, 5*time.Second)
			if err != nil {
				log.Printf("Error dequeuing %s job: %v", jobType, err)
				time.Sleep(1 * time.Second)
				continue
			}

			if job == nil {
				// No job available, continue polling
				continue
			}

			// Process the job
			log.Printf("Processing %s job %s", jobType, job.JobID)
			if err := w.Process(ctx, job); err != nil {
				log.Printf("Error processing %s job %s: %v", jobType, job.JobID, err)
				// TODO: Update job status to failed in database
			}
		}
	}
}

// BaseWorker provides common functionality for workers
type BaseWorker struct {
	db    *db.DB
	queue *queue.Queue
}

// NewBaseWorker creates a new base worker
func NewBaseWorker(database *db.DB, q *queue.Queue) *BaseWorker {
	return &BaseWorker{
		db:    database,
		queue: q,
	}
}

// UpdateJobProgress updates the job progress in the database
func (w *BaseWorker) UpdateJobProgress(ctx context.Context, jobID string, progress int) error {
	// TODO: Implement database update
	log.Printf("Job %s progress: %d%%", jobID, progress)
	return nil
}

// MarkJobDone marks a job as completed
func (w *BaseWorker) MarkJobDone(ctx context.Context, jobID string, output interface{}) error {
	// TODO: Implement database update
	log.Printf("Job %s completed", jobID)
	return nil
}

// MarkJobFailed marks a job as failed
func (w *BaseWorker) MarkJobFailed(ctx context.Context, jobID string, err error) error {
	// TODO: Implement database update
	log.Printf("Job %s failed: %v", jobID, err)
	return nil
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 2 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}
}
