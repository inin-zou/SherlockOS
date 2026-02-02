package queue

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/models"
)

// MemoryQueue is an in-memory queue implementation for development/testing
type MemoryQueue struct {
	queues map[string]chan *JobMessage
	mu     sync.RWMutex
}

// NewMemoryQueue creates a new in-memory queue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		queues: make(map[string]chan *JobMessage),
	}
}

// Close is a no-op for memory queue
func (q *MemoryQueue) Close() error {
	return nil
}

// getOrCreateQueue gets or creates a channel for the given job type
func (q *MemoryQueue) getOrCreateQueue(jobType models.JobType) chan *JobMessage {
	queueName := GetQueueName(jobType)

	q.mu.Lock()
	defer q.mu.Unlock()

	if ch, ok := q.queues[queueName]; ok {
		return ch
	}

	ch := make(chan *JobMessage, 1000) // Buffer for 1000 jobs
	q.queues[queueName] = ch
	return ch
}

// Enqueue adds a job to the queue
func (q *MemoryQueue) Enqueue(ctx context.Context, job *models.Job) error {
	ch := q.getOrCreateQueue(job.Type)

	msg := &JobMessage{
		JobID:      job.ID,
		CaseID:     job.CaseID,
		Type:       job.Type,
		Input:      job.Input,
		EnqueuedAt: time.Now().UTC(),
		Attempts:   0,
	}

	select {
	case ch <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue full, but don't block
		return nil
	}
}

// Dequeue retrieves a job from the queue
func (q *MemoryQueue) Dequeue(ctx context.Context, jobType models.JobType, timeout time.Duration) (*JobMessage, error) {
	ch := q.getOrCreateQueue(jobType)

	select {
	case msg := <-ch:
		msg.Attempts++
		now := time.Now().UTC()
		msg.LastAttempt = &now
		return msg, nil
	case <-time.After(timeout):
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Ack acknowledges successful job completion (no-op for memory queue)
func (q *MemoryQueue) Ack(ctx context.Context, msg *JobMessage) error {
	return nil
}

// Nack returns a failed job to the queue for retry
func (q *MemoryQueue) Nack(ctx context.Context, msg *JobMessage, maxRetries int) error {
	if msg.Attempts >= maxRetries {
		return nil // Just drop it for memory queue
	}

	// Re-queue
	ch := q.getOrCreateQueue(msg.Type)

	select {
	case ch <- msg:
		return nil
	default:
		return nil // Queue full, drop it
	}
}

// QueueLength returns the number of jobs in a queue
func (q *MemoryQueue) QueueLength(ctx context.Context, jobType models.JobType) (int64, error) {
	ch := q.getOrCreateQueue(jobType)
	return int64(len(ch)), nil
}

// ProcessingLength returns 0 for memory queue (no separate processing queue)
func (q *MemoryQueue) ProcessingLength(ctx context.Context, jobType models.JobType) (int64, error) {
	return 0, nil
}

// DLQLength returns 0 for memory queue (no DLQ)
func (q *MemoryQueue) DLQLength(ctx context.Context, jobType models.JobType) (int64, error) {
	return 0, nil
}

// RecoverStaleJobs is a no-op for memory queue
func (q *MemoryQueue) RecoverStaleJobs(ctx context.Context, jobType models.JobType) (int, error) {
	return 0, nil
}

// JobQueue interface that both Queue and MemoryQueue implement
type JobQueue interface {
	Enqueue(ctx context.Context, job *models.Job) error
	Dequeue(ctx context.Context, jobType models.JobType, timeout time.Duration) (*JobMessage, error)
	Ack(ctx context.Context, msg *JobMessage) error
	Nack(ctx context.Context, msg *JobMessage, maxRetries int) error
	QueueLength(ctx context.Context, jobType models.JobType) (int64, error)
	RecoverStaleJobs(ctx context.Context, jobType models.JobType) (int, error)
	Close() error
}

// Ensure both types implement JobQueue
var _ JobQueue = (*Queue)(nil)
var _ JobQueue = (*MemoryQueue)(nil)

// NewWithFallback creates a Redis queue, falling back to memory queue if Redis is unavailable
func NewWithFallback(redisURL string) (JobQueue, error) {
	if redisURL == "" {
		return NewMemoryQueue(), nil
	}

	q, err := New(redisURL)
	if err != nil {
		// Fall back to memory queue
		return NewMemoryQueue(), nil
	}

	return q, nil
}

// Helper to create JobMessage from a Job (for compatibility with raw job inputs)
func NewJobMessage(jobID, caseID uuid.UUID, jobType models.JobType, input interface{}) (*JobMessage, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	return &JobMessage{
		JobID:      jobID,
		CaseID:     caseID,
		Type:       jobType,
		Input:      inputJSON,
		EnqueuedAt: time.Now().UTC(),
		Attempts:   0,
	}, nil
}
