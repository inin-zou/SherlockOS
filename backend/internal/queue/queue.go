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

	// Default visibility timeout (how long a job is hidden after dequeue)
	DefaultVisibilityTimeout = 5 * time.Minute
)

// Queue manages job queuing with Redis
type Queue struct {
	client *redis.Client
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

	return &Queue{client: client}, nil
}

// Close closes the Redis connection
func (q *Queue) Close() error {
	return q.client.Close()
}

// JobMessage represents a job in the queue
type JobMessage struct {
	JobID     uuid.UUID       `json:"job_id"`
	CaseID    uuid.UUID       `json:"case_id"`
	Type      models.JobType  `json:"type"`
	Input     json.RawMessage `json:"input"`
	EnqueuedAt time.Time      `json:"enqueued_at"`
}

// Enqueue adds a job to the appropriate queue
func (q *Queue) Enqueue(ctx context.Context, job *models.Job) error {
	queueName := getQueueName(job.Type)

	msg := JobMessage{
		JobID:      job.ID,
		CaseID:     job.CaseID,
		Type:       job.Type,
		Input:      job.Input,
		EnqueuedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal job message: %w", err)
	}

	// Use LPUSH for FIFO with BRPOP
	if err := q.client.LPush(ctx, queueName, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// Dequeue retrieves a job from the queue (blocking)
func (q *Queue) Dequeue(ctx context.Context, jobType models.JobType, timeout time.Duration) (*JobMessage, error) {
	queueName := getQueueName(jobType)

	// BRPOP blocks until an item is available or timeout
	result, err := q.client.BRPop(ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Timeout, no job available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	// result[0] is queue name, result[1] is the value
	var msg JobMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job message: %w", err)
	}

	return &msg, nil
}

// QueueLength returns the number of jobs in a queue
func (q *Queue) QueueLength(ctx context.Context, jobType models.JobType) (int64, error) {
	queueName := getQueueName(jobType)
	return q.client.LLen(ctx, queueName).Result()
}

// getQueueName returns the queue name for a job type
func getQueueName(jobType models.JobType) string {
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
