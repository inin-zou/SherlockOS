package workers

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

// ErrorType categorizes errors for retry decisions
type ErrorType int

const (
	// ErrorTypeRetryable indicates the error is transient and can be retried
	ErrorTypeRetryable ErrorType = iota
	// ErrorTypeFatal indicates the error is permanent and should not be retried
	ErrorTypeFatal
)

// WorkerError wraps errors with type information
type WorkerError struct {
	Err  error
	Type ErrorType
}

func (e *WorkerError) Error() string {
	return e.Err.Error()
}

func (e *WorkerError) Unwrap() error {
	return e.Err
}

// NewRetryableError creates a retryable error
func NewRetryableError(err error) *WorkerError {
	return &WorkerError{Err: err, Type: ErrorTypeRetryable}
}

// NewFatalError creates a fatal error
func NewFatalError(err error) *WorkerError {
	return &WorkerError{Err: err, Type: ErrorTypeFatal}
}

// IsRetryable checks if an error should be retried
func IsRetryable(err error) bool {
	var workerErr *WorkerError
	if errors.As(err, &workerErr) {
		return workerErr.Type == ErrorTypeRetryable
	}
	// Default to retryable for unknown errors
	return true
}

// Worker interface defines the contract for all workers
type Worker interface {
	// Type returns the job type this worker handles
	Type() models.JobType

	// Process handles a single job
	Process(ctx context.Context, job *queue.JobMessage) error
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

// CalculateBackoff returns the backoff duration for a given attempt
func (c RetryConfig) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.InitialInterval
	}

	backoff := c.InitialInterval
	for i := 0; i < attempt-1; i++ {
		backoff = time.Duration(float64(backoff) * c.Multiplier)
		if backoff > c.MaxInterval {
			return c.MaxInterval
		}
	}
	return backoff
}

// Manager manages worker lifecycle
type Manager struct {
	repo        *db.Repository
	queue       queue.JobQueue
	workers     map[models.JobType]Worker
	retryConfig RetryConfig
	wg          sync.WaitGroup
	shutdown    chan struct{}

	// Heartbeat configuration
	heartbeatInterval time.Duration
	zombieTimeout     time.Duration
}

// ManagerConfig holds configuration for the manager
type ManagerConfig struct {
	RetryConfig       RetryConfig
	HeartbeatInterval time.Duration
	ZombieTimeout     time.Duration
}

// DefaultManagerConfig returns the default manager configuration
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		RetryConfig:       DefaultRetryConfig(),
		HeartbeatInterval: 30 * time.Second,
		ZombieTimeout:     2 * time.Minute,
	}
}

// NewManager creates a new worker manager
func NewManager(database *db.DB, q queue.JobQueue, config ManagerConfig) *Manager {
	var repo *db.Repository
	if database != nil {
		repo = db.NewRepository(database)
	}

	return &Manager{
		repo:              repo,
		queue:             q,
		workers:           make(map[models.JobType]Worker),
		retryConfig:       config.RetryConfig,
		heartbeatInterval: config.HeartbeatInterval,
		zombieTimeout:     config.ZombieTimeout,
		shutdown:          make(chan struct{}),
	}
}

// Register adds a worker for a specific job type
func (m *Manager) Register(w Worker) {
	m.workers[w.Type()] = w
	// Also register in the global registry so API handlers can validate requests
	GetGlobalRegistry().Register(w.Type())
}

// Start begins processing jobs for all registered workers
func (m *Manager) Start(ctx context.Context) {
	// Start zombie recovery goroutine
	m.wg.Add(1)
	go m.runZombieRecovery(ctx)

	// Start workers
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
			m.processJob(ctx, w, job)
		}
	}
}

// processJob handles a single job with heartbeat and error handling
func (m *Manager) processJob(ctx context.Context, w Worker, job *queue.JobMessage) {
	log.Printf("Processing %s job %s (attempt %d)", job.Type, job.JobID, job.Attempts)

	// Update job status to running
	if m.repo != nil {
		if err := m.repo.UpdateJobStatus(ctx, job.JobID, models.JobStatusRunning, 0); err != nil {
			log.Printf("Failed to update job status: %v", err)
		}
	}

	// Create a context with cancellation for heartbeat
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start heartbeat goroutine
	heartbeatDone := make(chan struct{})
	go m.runHeartbeat(jobCtx, job.JobID, heartbeatDone)

	// Process the job
	err := w.Process(jobCtx, job)

	// Stop heartbeat
	cancel()
	<-heartbeatDone

	if err != nil {
		log.Printf("Error processing %s job %s: %v", job.Type, job.JobID, err)
		m.handleJobError(ctx, job, err)
	} else {
		m.handleJobSuccess(ctx, job)
	}
}

// runHeartbeat updates the job timestamp periodically
func (m *Manager) runHeartbeat(ctx context.Context, jobID uuid.UUID, done chan struct{}) {
	defer close(done)

	ticker := time.NewTicker(m.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if m.repo != nil {
				if err := m.repo.UpdateJobHeartbeat(ctx, jobID); err != nil {
					log.Printf("Failed to update heartbeat for job %s: %v", jobID, err)
				}
			}
		}
	}
}

// handleJobSuccess handles successful job completion
func (m *Manager) handleJobSuccess(ctx context.Context, job *queue.JobMessage) {
	log.Printf("Job %s completed successfully", job.JobID)

	// Acknowledge the job in the queue
	if err := m.queue.Ack(ctx, job); err != nil {
		log.Printf("Failed to ack job %s: %v", job.JobID, err)
	}

	// Update job status to done (output should be set by worker)
	if m.repo != nil {
		if err := m.repo.UpdateJobStatus(ctx, job.JobID, models.JobStatusDone, 100); err != nil {
			log.Printf("Failed to update job status: %v", err)
		}
	}
}

// handleJobError handles job failures with retry logic
func (m *Manager) handleJobError(ctx context.Context, job *queue.JobMessage, err error) {
	if !IsRetryable(err) || job.Attempts >= m.retryConfig.MaxAttempts {
		// Fatal error or max retries exceeded
		log.Printf("Job %s failed permanently: %v", job.JobID, err)

		// Ack the job to remove it from the queue (don't requeue fatal errors)
		if queueErr := m.queue.Ack(ctx, job); queueErr != nil {
			log.Printf("Failed to ack failed job %s: %v", job.JobID, queueErr)
		}

		// Update job status to failed
		if m.repo != nil {
			if updateErr := m.repo.UpdateJobError(ctx, job.JobID, err.Error()); updateErr != nil {
				log.Printf("Failed to update job error: %v", updateErr)
			}
		}
		return
	}

	// Calculate backoff
	backoff := m.retryConfig.CalculateBackoff(job.Attempts)
	log.Printf("Job %s will retry after %v (attempt %d/%d)", job.JobID, backoff, job.Attempts, m.retryConfig.MaxAttempts)

	// Nack the job (will be requeued)
	if queueErr := m.queue.Nack(ctx, job, m.retryConfig.MaxAttempts); queueErr != nil {
		log.Printf("Failed to nack job %s: %v", job.JobID, queueErr)
	}

	// Update job status back to queued
	if m.repo != nil {
		if updateErr := m.repo.UpdateJobStatus(ctx, job.JobID, models.JobStatusQueued, 0); updateErr != nil {
			log.Printf("Failed to update job status: %v", updateErr)
		}
	}

	// Sleep for backoff duration before the job can be picked up again
	// Note: In production, you might want to use a scheduled job instead
	time.Sleep(backoff)
}

// runZombieRecovery periodically checks for and recovers zombie jobs
func (m *Manager) runZombieRecovery(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.zombieTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-m.shutdown:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.recoverZombieJobs(ctx)
		}
	}
}

// recoverZombieJobs finds and recovers stuck jobs
func (m *Manager) recoverZombieJobs(ctx context.Context) {
	// Recover from queue processing queues
	for jobType := range m.workers {
		recovered, err := m.queue.RecoverStaleJobs(ctx, jobType)
		if err != nil {
			log.Printf("Failed to recover stale %s jobs from queue: %v", jobType, err)
		} else if recovered > 0 {
			log.Printf("Recovered %d stale %s jobs from processing queue", recovered, jobType)
		}
	}

	// Also check database for zombie jobs
	if m.repo != nil {
		zombies, err := m.repo.GetZombieJobs(ctx, m.zombieTimeout)
		if err != nil {
			log.Printf("Failed to get zombie jobs from database: %v", err)
			return
		}

		for _, job := range zombies {
			log.Printf("Found zombie job %s, requeueing", job.ID)

			// Increment retry count
			requeued, err := m.repo.IncrementJobRetry(ctx, job.ID, m.retryConfig.MaxAttempts)
			if err != nil {
				log.Printf("Failed to increment retry for zombie job %s: %v", job.ID, err)
				continue
			}

			if requeued {
				// Re-enqueue the job
				if err := m.queue.Enqueue(ctx, job); err != nil {
					log.Printf("Failed to requeue zombie job %s: %v", job.ID, err)
				}
			} else {
				log.Printf("Zombie job %s exceeded max retries, marked as failed", job.ID)
			}
		}
	}
}

// BaseWorker provides common functionality for workers
type BaseWorker struct {
	repo  *db.Repository
	queue queue.JobQueue
}

// NewBaseWorker creates a new base worker
func NewBaseWorker(database *db.DB, q queue.JobQueue) *BaseWorker {
	var repo *db.Repository
	if database != nil {
		repo = db.NewRepository(database)
	}
	return &BaseWorker{
		repo:  repo,
		queue: q,
	}
}

// UpdateJobProgress updates the job progress in the database
func (w *BaseWorker) UpdateJobProgress(ctx context.Context, jobID uuid.UUID, progress int) error {
	if w.repo == nil {
		log.Printf("Job %s progress: %d%% (no db)", jobID, progress)
		return nil
	}

	if err := w.repo.UpdateJobStatus(ctx, jobID, models.JobStatusRunning, progress); err != nil {
		return err
	}
	log.Printf("Job %s progress: %d%%", jobID, progress)
	return nil
}

// MarkJobDone marks a job as completed with output
func (w *BaseWorker) MarkJobDone(ctx context.Context, jobID uuid.UUID, output interface{}) error {
	if w.repo == nil {
		log.Printf("Job %s completed (no db)", jobID)
		return nil
	}

	if err := w.repo.UpdateJobOutput(ctx, jobID, output); err != nil {
		return err
	}
	log.Printf("Job %s completed", jobID)
	return nil
}

// MarkJobFailed marks a job as failed with error message
func (w *BaseWorker) MarkJobFailed(ctx context.Context, jobID uuid.UUID, err error) error {
	if w.repo == nil {
		log.Printf("Job %s failed: %v (no db)", jobID, err)
		return nil
	}

	if updateErr := w.repo.UpdateJobError(ctx, jobID, err.Error()); updateErr != nil {
		return updateErr
	}
	log.Printf("Job %s failed: %v", jobID, err)
	return nil
}
