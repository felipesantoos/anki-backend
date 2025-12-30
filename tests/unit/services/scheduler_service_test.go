package services

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/services/jobs"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerService_Schedule(t *testing.T) {
	mockScheduler := new(MockJobScheduler)
	service := jobs.NewSchedulerService(mockScheduler)

	cron := "0 0 * * * *"
	jobType := "test-job"
	payload := map[string]interface{}{"key": "value"}

	t.Run("Success", func(t *testing.T) {
		mockScheduler.On("Schedule", cron, jobType, payload).Return(nil).Once()

		err := service.Schedule(cron, jobType, payload)

		assert.NoError(t, err)
		mockScheduler.AssertExpectations(t)
	})

	t.Run("EmptyCron", func(t *testing.T) {
		err := service.Schedule("", jobType, payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cron expression cannot be empty")
	})

	t.Run("EmptyJobType", func(t *testing.T) {
		err := service.Schedule(cron, "", payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job type cannot be empty")
	})
}

func TestSchedulerService_StartStop(t *testing.T) {
	mockScheduler := new(MockJobScheduler)
	service := jobs.NewSchedulerService(mockScheduler)

	t.Run("Start", func(t *testing.T) {
		mockScheduler.On("Start").Once()
		service.Start()
		mockScheduler.AssertExpectations(t)
	})

	t.Run("Stop", func(t *testing.T) {
		mockScheduler.On("Stop").Once()
		service.Stop()
		mockScheduler.AssertExpectations(t)
	})
}

