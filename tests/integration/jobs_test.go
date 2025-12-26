package integration

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/jobs"
	"github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// testHandler is a simple handler for testing
type testHandler struct {
	jobType string
	handled chan *secondary.Job
}

func NewTestHandler(jobType string) *testHandler {
	return &testHandler{
		jobType: jobType,
		handled: make(chan *secondary.Job, 10),
	}
}

func (h *testHandler) Handle(ctx context.Context, job *secondary.Job) error {
	select {
	case h.handled <- job:
	default:
	}
	return nil
}

func (h *testHandler) JobType() string {
	return h.jobType
}

func TestJobs_EnqueueAndProcess(t *testing.T) {
	// Check if Redis is available
	redisCfg := config.RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     getEnvOrDefault("REDIS_PORT", "6380"), // Use test Redis port
		Password: "",
		DB:       0,
	}

	log := logger.GetLogger()
	if log == nil {
		logger.InitLogger("INFO", "development")
		log = logger.GetLogger()
	}

	rdb, err := redis.NewRedisRepository(redisCfg, log)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer rdb.Close()

	// Create job queue
	queue := jobs.NewRedisQueue(rdb.Client, "test:jobs:queue")

	// Create registry
	registry := jobs.NewJobRegistry()

	// Create test handler
	testHandler := NewTestHandler("test_job")
	if err := registry.Register(testHandler); err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Create worker pool
	workerPool := jobs.NewWorkerPool(2, queue, registry, log, 3, 1)
	workerPool.Start()
	defer workerPool.Stop()

	// Enqueue a job
	ctx := context.Background()
	testJob := jobs.NewJob("test_job", map[string]interface{}{"key": "value"}, 3)

	err = queue.Enqueue(ctx, testJob)
	if err != nil {
		t.Fatalf("Failed to enqueue job: %v", err)
	}

	// Wait for job to be processed (with timeout)
	select {
	case handledJob := <-testHandler.handled:
		if handledJob.ID != testJob.ID {
			t.Errorf("Expected job ID %s, got %s", testJob.ID, handledJob.ID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Job was not processed within timeout")
	}

	// Wait a bit for status to be updated (race condition fix)
	time.Sleep(100 * time.Millisecond)

	// Verify job status with retries
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		statusJob, err := queue.GetStatus(ctx, testJob.ID)
		if err != nil {
			// Job status might not be stored yet, wait and retry
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if statusJob.Status == secondary.JobStatusCompleted {
			return // Success
		}

		if statusJob.Status == secondary.JobStatusFailed {
			t.Errorf("Job failed unexpectedly: %s", statusJob.Error)
			return
		}

		// Status is still pending or processing, wait a bit more
		time.Sleep(100 * time.Millisecond)
	}

	// Final check
	statusJob, err := queue.GetStatus(ctx, testJob.ID)
	if err != nil {
		t.Fatalf("Failed to get job status: %v", err)
	}

	if statusJob.Status != secondary.JobStatusCompleted {
		t.Errorf("Expected job status %s, got %s", secondary.JobStatusCompleted, statusJob.Status)
	}
}

func TestJobs_RetryLogic(t *testing.T) {
	redisCfg := config.RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     getEnvOrDefault("REDIS_PORT", "6380"),
		Password: "",
		DB:       0,
	}

	log := logger.GetLogger()
	if log == nil {
		logger.InitLogger("INFO", "development")
		log = logger.GetLogger()
	}

	rdb, err := redis.NewRedisRepository(redisCfg, log)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer rdb.Close()

	queue := jobs.NewRedisQueue(rdb.Client, "test:jobs:retry")
	registry := jobs.NewJobRegistry()

	// Handler that fails first two times, then succeeds
	var attempts int
	failHandler := &failingHandler{
		jobType:  "failing_job",
		attempts: &attempts,
		failUntil: 2,
	}
	if err := registry.Register(failHandler); err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	workerPool := jobs.NewWorkerPool(1, queue, registry, log, 3, 1)
	workerPool.Start()
	defer workerPool.Stop()

	// Enqueue a job
	ctx := context.Background()
	testJob := jobs.NewJob("failing_job", map[string]interface{}{}, 3)

	err = queue.Enqueue(ctx, testJob)
	if err != nil {
		t.Fatalf("Failed to enqueue job: %v", err)
	}

	// Wait for job to eventually succeed (with timeout)
	// The job should fail twice (attempts 1 and 2), then succeed on attempt 3
	// But due to retry logic, retries are incremented before the handler is called again
	// So we expect: attempt 1 (retries=0) fails -> retry (retries=1) -> attempt 2 (retries=1) fails -> retry (retries=2) -> attempt 3 (retries=2) succeeds
	timeout := time.After(15 * time.Second) // Increased timeout to account for retry delays
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Check final status before failing
			statusJob, err := queue.GetStatus(ctx, testJob.ID)
			if err == nil {
				t.Fatalf("Job did not complete within timeout. Final status: %s, attempts: %d, retries: %d", 
					statusJob.Status, attempts, statusJob.Retries)
			} else {
				t.Fatalf("Job did not complete within timeout. Status unavailable. Attempts: %d", attempts)
			}
		case <-ticker.C:
			statusJob, err := queue.GetStatus(ctx, testJob.ID)
			if err != nil {
				continue
			}
			if statusJob.Status == secondary.JobStatusCompleted {
				// Verify that the job was retried at least once
				// The handler should have been called at least 2 times (fail, fail, succeed)
				if attempts < 2 {
					t.Errorf("Expected at least 2 attempts, got %d", attempts)
				}
				return // Success
			}
			if statusJob.Status == secondary.JobStatusFailed {
				t.Fatalf("Job failed permanently. Error: %s, Attempts: %d, Retries: %d", 
					statusJob.Error, attempts, statusJob.Retries)
			}
		}
	}
}

type failingHandler struct {
	jobType   string
	attempts  *int
	failUntil int
}

func (h *failingHandler) Handle(ctx context.Context, job *secondary.Job) error {
	*h.attempts++
	if *h.attempts <= h.failUntil {
		return errors.New("simulated failure")
	}
	return nil
}

func (h *failingHandler) JobType() string {
	return h.jobType
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

