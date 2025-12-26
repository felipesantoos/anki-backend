package jobs

import (
	"context"
	"fmt"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/jobs"
)

// JobService provides high-level operations for job management
type JobService struct {
	queue      secondary.IJobQueue
	maxRetries int
}

// NewJobService creates a new job service
func NewJobService(queue secondary.IJobQueue, maxRetries int) *JobService {
	return &JobService{
		queue:      queue,
		maxRetries: maxRetries,
	}
}

// Enqueue adds a job to the queue
func (s *JobService) Enqueue(ctx context.Context, jobType string, payload map[string]interface{}) (string, error) {
	if jobType == "" {
		return "", fmt.Errorf("job type cannot be empty")
	}

	job := jobs.NewJob(jobType, payload, s.maxRetries)

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	return job.ID, nil
}

// EnqueueWithRetries adds a job to the queue with custom max retries
func (s *JobService) EnqueueWithRetries(ctx context.Context, jobType string, payload map[string]interface{}, maxRetries int) (string, error) {
	if jobType == "" {
		return "", fmt.Errorf("job type cannot be empty")
	}

	if maxRetries < 0 {
		return "", fmt.Errorf("max retries cannot be negative")
	}

	job := NewJob(jobType, payload, maxRetries)

	if err := s.queue.Enqueue(ctx, job); err != nil {
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	return job.ID, nil
}

// GetStatus retrieves the status of a job by ID
func (s *JobService) GetStatus(ctx context.Context, jobID string) (*secondary.Job, error) {
	if jobID == "" {
		return nil, fmt.Errorf("job ID cannot be empty")
	}

	job, err := s.queue.GetStatus(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	return job, nil
}

