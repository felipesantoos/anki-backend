package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// WorkerPool manages a pool of workers that process jobs from a queue
type WorkerPool struct {
	workers      int
	queue        secondary.IJobQueue
	registry     *JobRegistry
	logger       *slog.Logger
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownCh   chan struct{}
	maxRetries   int
	retryDelay   int // in seconds
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(
	workers int,
	queue secondary.IJobQueue,
	registry *JobRegistry,
	logger *slog.Logger,
	maxRetries int,
	retryDelaySeconds int,
) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:    workers,
		queue:      queue,
		registry:   registry,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		shutdownCh: make(chan struct{}),
		maxRetries: maxRetries,
		retryDelay: retryDelaySeconds,
	}
}

// Start starts all workers in the pool
func (wp *WorkerPool) Start() {
	wp.logger.Info("Starting worker pool",
		"workers", wp.workers,
		"max_retries", wp.maxRetries,
		"retry_delay_seconds", wp.retryDelay,
	)

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop stops all workers gracefully
// Waits for currently processing jobs to complete
func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool...")

	// Cancel context to signal workers to stop
	wp.cancel()

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		wp.logger.Info("Worker pool stopped successfully")
	case <-time.After(30 * time.Second):
		wp.logger.Warn("Worker pool stop timeout reached, some workers may still be running")
	}

	close(wp.shutdownCh)
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker started", "worker_id", id)

	dequeueTimeout := 5 * time.Second

	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Debug("Worker stopping", "worker_id", id)
			return
		default:
			// Dequeue a job
			job, err := wp.queue.Dequeue(wp.ctx, dequeueTimeout)
			if err != nil {
				// Timeout is expected when queue is empty
				if err.Error() == "no job available within timeout" {
					continue
				}
				wp.logger.Error("Failed to dequeue job",
					"worker_id", id,
					"error", err,
				)
				time.Sleep(1 * time.Second) // Brief pause before retry
				continue
			}

			// Process the job
			wp.processJob(id, job)
		}
	}
}

// processJob processes a single job
func (wp *WorkerPool) processJob(workerID int, job *secondary.Job) {
	wp.logger.Info("Processing job",
		"worker_id", workerID,
		"job_id", job.ID,
		"job_type", job.Type,
		"retries", job.Retries,
	)

	// Get handler for job type
	handler, err := wp.registry.GetHandler(job.Type)
	if err != nil {
		wp.logger.Error("No handler found for job type",
			"worker_id", workerID,
			"job_id", job.ID,
			"job_type", job.Type,
			"error", err,
		)
		wp.handleJobFailure(job, fmt.Sprintf("no handler found: %v", err))
		return
	}

	// Create context with timeout for job processing
	jobCtx, cancel := context.WithTimeout(wp.ctx, 10*time.Minute)
	defer cancel()

	// Process the job
	err = handler.Handle(jobCtx, job)
	if err != nil {
		wp.logger.Warn("Job processing failed",
			"worker_id", workerID,
			"job_id", job.ID,
			"job_type", job.Type,
			"error", err,
			"retries", job.Retries,
		)

		// Check if job should be retried
		if ShouldRetry(job) {
			wp.retryJob(job, err)
		} else {
			wp.handleJobFailure(job, err.Error())
		}
		return
	}

	// Job completed successfully
	wp.handleJobSuccess(job)
}

// retryJob retries a failed job
func (wp *WorkerPool) retryJob(job *secondary.Job, err error) {
	IncrementRetry(job)

	// Calculate retry delay (exponential backoff)
	delay := CalculateRetryDelay(wp.retryDelay, job.Retries)

	wp.logger.Info("Retrying job",
		"job_id", job.ID,
		"job_type", job.Type,
		"retry_count", job.Retries,
		"max_retries", job.MaxRetries,
		"delay_seconds", delay.Seconds(),
		"error", err,
	)

	// Store error in job
	job.Error = err.Error()

	// Update status to pending
	job.Status = secondary.JobStatusPending
	job.ProcessedAt = nil

	// Re-enqueue after delay (using goroutine to not block worker)
	go func() {
		time.Sleep(delay)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := wp.queue.Retry(ctx, job); err != nil {
			wp.logger.Error("Failed to retry job",
				"job_id", job.ID,
				"error", err,
			)
		}
	}()
}

// handleJobSuccess handles a successfully completed job
func (wp *WorkerPool) handleJobSuccess(job *secondary.Job) {
	job.Status = secondary.JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := wp.queue.UpdateStatus(ctx, job.ID, secondary.JobStatusCompleted); err != nil {
		wp.logger.Error("Failed to update job status to completed",
			"job_id", job.ID,
			"error", err,
		)
		return
	}

	wp.logger.Info("Job completed successfully",
		"job_id", job.ID,
		"job_type", job.Type,
		"retries", job.Retries,
	)
}

// handleJobFailure handles a permanently failed job
func (wp *WorkerPool) handleJobFailure(job *secondary.Job, errorMsg string) {
	job.Status = secondary.JobStatusFailed
	now := time.Now()
	job.FailedAt = &now
	job.Error = errorMsg

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := wp.queue.UpdateStatus(ctx, job.ID, secondary.JobStatusFailed); err != nil {
		wp.logger.Error("Failed to update job status to failed",
			"job_id", job.ID,
			"error", err,
		)
		return
	}

	wp.logger.Error("Job failed permanently",
		"job_id", job.ID,
		"job_type", job.Type,
		"retries", job.Retries,
		"max_retries", job.MaxRetries,
		"error", errorMsg,
	)
}

