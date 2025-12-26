package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// Scheduler implements IJobScheduler using cron expressions
type Scheduler struct {
	cron    *cron.Cron
	queue   secondary.IJobQueue
	logger  *slog.Logger
	ctx     context.Context
	enabled bool
}

// NewScheduler creates a new job scheduler
func NewScheduler(queue secondary.IJobQueue, logger *slog.Logger) *Scheduler {
	// Create cron with seconds support (5 fields: second, minute, hour, day, month)
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron:   c,
		queue:  queue,
		logger: logger,
		ctx:    context.Background(),
		enabled: false,
	}
}

// Schedule adds a recurring job using a cron expression
// Cron format: "second minute hour day month weekday"
// Examples:
//   - "0 0 2 * * *" - Daily at 2 AM
//   - "0 */5 * * * *" - Every 5 minutes
//   - "0 0 * * * *" - Every hour
func (s *Scheduler) Schedule(cronExpr string, jobType string, payload map[string]interface{}) error {
	_, err := s.cron.AddFunc(cronExpr, func() {
		if !s.enabled {
			return
		}

		// Create a new job
		job := NewJob(jobType, payload, 3) // Default max retries: 3

		// Enqueue the job
		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
		defer cancel()

		if err := s.queue.Enqueue(ctx, job); err != nil {
			s.logger.Error("Failed to enqueue scheduled job",
				"cron_expr", cronExpr,
				"job_type", jobType,
				"error", err,
			)
			return
		}

		s.logger.Info("Scheduled job enqueued",
			"cron_expr", cronExpr,
			"job_type", jobType,
			"job_id", job.ID,
		)
	})

	if err != nil {
		return err
	}

	s.logger.Info("Scheduled job registered",
		"cron_expr", cronExpr,
		"job_type", jobType,
	)

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.enabled = true
	s.cron.Start()
	s.logger.Info("Job scheduler started")
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() {
	s.enabled = false
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("Job scheduler stopped")
}

