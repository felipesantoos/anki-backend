package jobs

import (
	"fmt"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// SchedulerService provides high-level operations for scheduling jobs
type SchedulerService struct {
	scheduler secondary.IJobScheduler
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(scheduler secondary.IJobScheduler) *SchedulerService {
	return &SchedulerService{
		scheduler: scheduler,
	}
}

// Schedule registers a recurring job using a cron expression
// Cron format: "second minute hour day month weekday"
// Examples:
//   - "0 0 2 * * *" - Daily at 2 AM
//   - "0 */5 * * * *" - Every 5 minutes
//   - "0 0 * * * *" - Every hour
func (s *SchedulerService) Schedule(cronExpr string, jobType string, payload map[string]interface{}) error {
	if cronExpr == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}

	if jobType == "" {
		return fmt.Errorf("job type cannot be empty")
	}

	return s.scheduler.Schedule(cronExpr, jobType, payload)
}

// Start starts the scheduler
func (s *SchedulerService) Start() {
	s.scheduler.Start()
}

// Stop stops the scheduler gracefully
func (s *SchedulerService) Stop() {
	s.scheduler.Stop()
}

