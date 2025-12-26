package secondary

import (
	"context"
	"time"
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Status      JobStatus              `json:"status"`
	Retries     int                    `json:"retries"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// IJobQueue defines the interface for job queue operations
// Implementation agnostic - works with Redis, in-memory, etc.
type IJobQueue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job *Job) error

	// Dequeue removes and returns the next job from the queue
	// Blocks until a job is available or timeout is reached
	Dequeue(ctx context.Context, timeout time.Duration) (*Job, error)

	// GetStatus retrieves the status of a job by ID
	GetStatus(ctx context.Context, jobID string) (*Job, error)

	// UpdateStatus updates the status of a job
	UpdateStatus(ctx context.Context, jobID string, status JobStatus) error

	// Retry re-enqueues a failed job for retry
	Retry(ctx context.Context, job *Job) error
}

// IJobScheduler defines the interface for scheduling recurring jobs (cron)
type IJobScheduler interface {
	// Schedule adds a recurring job using a cron expression
	Schedule(cronExpr string, jobType string, payload map[string]interface{}) error

	// Start starts the scheduler
	Start()

	// Stop stops the scheduler gracefully
	Stop()
}

