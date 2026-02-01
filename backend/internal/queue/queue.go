package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/sherlockos/backend/internal/models"
)

const (
	// Queue names
	QueueReconstruction = "jobs:reconstruction"
	QueueImageGen       = "jobs:imagegen"
	QueueReasoning      = "jobs:reasoning"
	QueueProfile        = "jobs:profile"
	QueueExport         = "jobs:export"

	// Dead letter queue suffix
	DeadLetterSuffix = ":dlq"

	// Processing queue suffix (for visibility timeout)
	ProcessingSuffix = ":processing"

	// Default visibility timeout (how long a job is hidden after dequeue)
	DefaultVisibilityTimeout = 5 * time.Minute
)

// Queue manages job queuing with Redis
type Queue struct {
	client            *redis.Client
	visibilityTimeout time.Duration
}

// New creates a new Queue instance
func New(redisURL string) (*Queue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Queue{
		client:            client,
		visibilityTimeout: DefaultVisibilityTimeout,
	}, nil
}

// Close closes the Redis connection
func (q *Queue) Close() error {
	return q.client.Close()
}

// JobMessage represents a job in the queue
type JobMessage struct {
	JobID       uuid.UUID       `json:"job_id"`
	CaseID      uuid.UUID       `json:"case_id"`
	Type        models.JobType  `json:"type"`
	Input       json.RawMessage `json:"input"`
	EnqueuedAt  time.Time       `json:"enqueued_at"`
	Attempts    int             `json:"attempts"`
	LastAttempt *time.Time      `json:"last_attempt,omitempty"`
}

// Enqueue adds a job to the appropriate queue
func (q *Queue) Enqueue(ctx context.Context, job *models.Job) error {
	queueName := GetQueueName(job.Type)

	msg := JobMessage{
		JobID:      job.ID,
		CaseID:     job.CaseID,
		Type:       job.Type,
		Input:      job.Input,
		EnqueuedAt: time.Now().UTC(),
		Attempts:   0,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal job message: %w", err)
	}

	// Use LPUSH for FIFO with BRPOPLPUSH
	if err := q.client.LPush(ctx, queueName, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// Dequeue retrieves a job from the queue with visibility timeout
// The job is moved to a processing queue and must be acknowledged
func (q *Queue) Dequeue(ctx context.Context, jobType models.JobType, timeout time.Duration) (*JobMessage, error) {
	queueName := GetQueueName(jobType)
	processingQueue := queueName + ProcessingSuffix

	// Use BRPOPLPUSH to atomically move job to processing queue
	// This provides "at least once" delivery semantics
	result, err := q.client.BRPopLPush(ctx, queueName, processingQueue, timeout).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Timeout, no job available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	var msg JobMessage
	if err := json.Unmarshal([]byte(result), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job message: %w", err)
	}

	// Update attempt count
	msg.Attempts++
	now := time.Now().UTC()
	msg.LastAttempt = &now

	return &msg, nil
}

// Ack acknowledges successful job completion (removes from processing queue)
func (q *Queue) Ack(ctx context.Context, msg *JobMessage) error {
	processingQueue := GetQueueName(msg.Type) + ProcessingSuffix

	// Serialize the original message to find it in the list
	data, err := json.Marshal(JobMessage{
		JobID:      msg.JobID,
		CaseID:     msg.CaseID,
		Type:       msg.Type,
		Input:      msg.Input,
		EnqueuedAt: msg.EnqueuedAt,
		Attempts:   msg.Attempts - 1, // Original attempts before this processing
	})
	if err != nil {
		return fmt.Errorf("failed to marshal job message for ack: %w", err)
	}

	// Remove from processing queue
	if err := q.client.LRem(ctx, processingQueue, 1, data).Err(); err != nil {
		return fmt.Errorf("failed to ack job: %w", err)
	}

	return nil
}

// Nack returns a failed job to the queue for retry or moves to DLQ
func (q *Queue) Nack(ctx context.Context, msg *JobMessage, maxRetries int) error {
	processingQueue := GetQueueName(msg.Type) + ProcessingSuffix

	// Serialize the original message to find it
	originalData, err := json.Marshal(JobMessage{
		JobID:      msg.JobID,
		CaseID:     msg.CaseID,
		Type:       msg.Type,
		Input:      msg.Input,
		EnqueuedAt: msg.EnqueuedAt,
		Attempts:   msg.Attempts - 1,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal job message for nack: %w", err)
	}

	// Remove from processing queue
	if err := q.client.LRem(ctx, processingQueue, 1, originalData).Err(); err != nil {
		return fmt.Errorf("failed to remove from processing queue: %w", err)
	}

	// Check if max retries exceeded
	if msg.Attempts >= maxRetries {
		// Move to dead letter queue
		return q.moveToDLQ(ctx, msg)
	}

	// Re-queue for retry
	return q.requeue(ctx, msg)
}

// moveToDLQ moves a job to the dead letter queue
func (q *Queue) moveToDLQ(ctx context.Context, msg *JobMessage) error {
	dlqName := GetQueueName(msg.Type) + DeadLetterSuffix

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal job message for DLQ: %w", err)
	}

	if err := q.client.LPush(ctx, dlqName, data).Err(); err != nil {
		return fmt.Errorf("failed to move job to DLQ: %w", err)
	}

	return nil
}

// requeue adds a job back to the main queue for retry
func (q *Queue) requeue(ctx context.Context, msg *JobMessage) error {
	queueName := GetQueueName(msg.Type)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal job message for requeue: %w", err)
	}

	// Add to the front of the queue (RPUSH for retry priority)
	if err := q.client.RPush(ctx, queueName, data).Err(); err != nil {
		return fmt.Errorf("failed to requeue job: %w", err)
	}

	return nil
}

// RecoverStaleJobs moves jobs from processing queue back to main queue
// if they've been in processing longer than visibility timeout
func (q *Queue) RecoverStaleJobs(ctx context.Context, jobType models.JobType) (int, error) {
	processingQueue := GetQueueName(jobType) + ProcessingSuffix
	mainQueue := GetQueueName(jobType)

	// Get all jobs in processing queue
	jobs, err := q.client.LRange(ctx, processingQueue, 0, -1).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get processing jobs: %w", err)
	}

	recovered := 0
	cutoff := time.Now().UTC().Add(-q.visibilityTimeout)

	for _, jobData := range jobs {
		var msg JobMessage
		if err := json.Unmarshal([]byte(jobData), &msg); err != nil {
			continue
		}

		// Check if job has been in processing too long
		if msg.LastAttempt != nil && msg.LastAttempt.Before(cutoff) {
			// Move back to main queue
			pipe := q.client.TxPipeline()
			pipe.LRem(ctx, processingQueue, 1, jobData)
			pipe.LPush(ctx, mainQueue, jobData)
			if _, err := pipe.Exec(ctx); err != nil {
				continue
			}
			recovered++
		}
	}

	return recovered, nil
}

// QueueLength returns the number of jobs in a queue
func (q *Queue) QueueLength(ctx context.Context, jobType models.JobType) (int64, error) {
	queueName := GetQueueName(jobType)
	return q.client.LLen(ctx, queueName).Result()
}

// ProcessingLength returns the number of jobs currently being processed
func (q *Queue) ProcessingLength(ctx context.Context, jobType models.JobType) (int64, error) {
	processingQueue := GetQueueName(jobType) + ProcessingSuffix
	return q.client.LLen(ctx, processingQueue).Result()
}

// DLQLength returns the number of jobs in the dead letter queue
func (q *Queue) DLQLength(ctx context.Context, jobType models.JobType) (int64, error) {
	dlqName := GetQueueName(jobType) + DeadLetterSuffix
	return q.client.LLen(ctx, dlqName).Result()
}

// GetQueueName returns the queue name for a job type
func GetQueueName(jobType models.JobType) string {
	switch jobType {
	case models.JobTypeReconstruction:
		return QueueReconstruction
	case models.JobTypeImageGen:
		return QueueImageGen
	case models.JobTypeReasoning:
		return QueueReasoning
	case models.JobTypeProfile:
		return QueueProfile
	case models.JobTypeExport:
		return QueueExport
	default:
		return "jobs:unknown"
	}
}
