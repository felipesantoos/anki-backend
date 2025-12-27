package jobs

import (
	"context"
	"fmt"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/jobs"
	"github.com/felipesantos/anki-backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	ctx, span := tracing.StartSpan(ctx, "job.enqueue",
		trace.WithAttributes(
			attribute.String("job.type", jobType),
			attribute.Int("job.max_retries", s.maxRetries),
		),
	)
	defer span.End()

	if jobType == "" {
		err := fmt.Errorf("job type cannot be empty")
		tracing.RecordError(span, err)
		return "", err
	}

	job := jobs.NewJob(jobType, payload, s.maxRetries)

	if err := s.queue.Enqueue(ctx, job); err != nil {
		tracing.RecordError(span, err)
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	span.SetAttributes(attribute.String("job.id", job.ID))
	return job.ID, nil
}

// EnqueueWithRetries adds a job to the queue with custom max retries
func (s *JobService) EnqueueWithRetries(ctx context.Context, jobType string, payload map[string]interface{}, maxRetries int) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "job.enqueue_with_retries",
		trace.WithAttributes(
			attribute.String("job.type", jobType),
			attribute.Int("job.max_retries", maxRetries),
		),
	)
	defer span.End()

	if jobType == "" {
		err := fmt.Errorf("job type cannot be empty")
		tracing.RecordError(span, err)
		return "", err
	}

	if maxRetries < 0 {
		err := fmt.Errorf("max retries cannot be negative")
		tracing.RecordError(span, err)
		return "", err
	}

	job := jobs.NewJob(jobType, payload, maxRetries)

	if err := s.queue.Enqueue(ctx, job); err != nil {
		tracing.RecordError(span, err)
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	span.SetAttributes(attribute.String("job.id", job.ID))
	return job.ID, nil
}

// GetStatus retrieves the status of a job by ID
func (s *JobService) GetStatus(ctx context.Context, jobID string) (*secondary.Job, error) {
	ctx, span := tracing.StartSpan(ctx, "job.get_status",
		trace.WithAttributes(attribute.String("job.id", jobID)),
	)
	defer span.End()

	if jobID == "" {
		err := fmt.Errorf("job ID cannot be empty")
		tracing.RecordError(span, err)
		return nil, err
	}

	job, err := s.queue.GetStatus(ctx, jobID)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	if job != nil {
		span.SetAttributes(
			attribute.String("job.type", job.Type),
			attribute.String("job.status", string(job.Status)),
		)
	}
	return job, nil
}

