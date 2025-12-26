package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// RedisQueue implements IJobQueue using Redis Lists
type RedisQueue struct {
	client        *redis.Client
	queueKey      string
	statusKey     string
	retryQueueKey string
}

// NewRedisQueue creates a new Redis-based job queue
func NewRedisQueue(client *redis.Client, queueKey string) *RedisQueue {
	return &RedisQueue{
		client:        client,
		queueKey:      queueKey,
		statusKey:     fmt.Sprintf("%s:status", queueKey),
		retryQueueKey: fmt.Sprintf("%s:retry", queueKey),
	}
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *secondary.Job) error {
	// Serialize job
	data, err := SerializeJob(job)
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	// Add to queue (LPUSH - adds to left side of list)
	if err := q.client.LPush(ctx, q.queueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Store job status
	if err := q.storeJobStatus(ctx, job); err != nil {
		// Log error but don't fail the enqueue operation
		// Status storage is for tracking, not critical for execution
	}

	return nil
}

// Dequeue removes and returns the next job from the queue
// Blocks until a job is available or timeout is reached
func (q *RedisQueue) Dequeue(ctx context.Context, timeout time.Duration) (*secondary.Job, error) {
	// Use BRPOP to block until a job is available
	// Blocks for up to timeout duration
	result, err := q.client.BRPop(ctx, timeout, q.queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no job available within timeout")
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid result from Redis: expected 2 elements, got %d", len(result))
	}

	// Deserialize job
	job, err := DeserializeJob([]byte(result[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	// Update status to processing
	job.Status = secondary.JobStatusProcessing
	now := time.Now()
	job.ProcessedAt = &now
	if err := q.storeJobStatus(ctx, job); err != nil {
		// Log error but continue processing
	}

	return job, nil
}

// GetStatus retrieves the status of a job by ID
func (q *RedisQueue) GetStatus(ctx context.Context, jobID string) (*secondary.Job, error) {
	key := q.getStatusKey(jobID)
	
	data, err := q.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	job, err := DeserializeJob([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	return job, nil
}

// UpdateStatus updates the status of a job
func (q *RedisQueue) UpdateStatus(ctx context.Context, jobID string, status secondary.JobStatus) error {
	job, err := q.GetStatus(ctx, jobID)
	if err != nil {
		return err
	}

	job.Status = status
	now := time.Now()
	
	switch status {
	case secondary.JobStatusCompleted:
		job.CompletedAt = &now
	case secondary.JobStatusFailed:
		job.FailedAt = &now
	}

	return q.storeJobStatus(ctx, job)
}

// Retry re-enqueues a failed job for retry
// Note: The retry count should already be incremented by the caller
func (q *RedisQueue) Retry(ctx context.Context, job *secondary.Job) error {
	// Reset job state for retry
	job.Status = secondary.JobStatusPending
	job.ProcessedAt = nil
	job.FailedAt = nil
	job.Error = ""

	// Re-enqueue
	return q.Enqueue(ctx, job)
}

// storeJobStatus stores job status in Redis with TTL
func (q *RedisQueue) storeJobStatus(ctx context.Context, job *secondary.Job) error {
	key := q.getStatusKey(job.ID)
	
	data, err := SerializeJob(job)
	if err != nil {
		return err
	}

	// Store with 24 hour TTL (jobs older than this are considered stale)
	ttl := 24 * time.Hour
	return q.client.Set(ctx, key, data, ttl).Err()
}

// getStatusKey returns the Redis key for a job status
func (q *RedisQueue) getStatusKey(jobID string) string {
	return fmt.Sprintf("%s:%s", q.statusKey, jobID)
}

