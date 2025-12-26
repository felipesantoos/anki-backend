package jobs

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// GenerateJobID generates a unique job ID
func GenerateJobID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("job_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// NewJob creates a new job with default values
func NewJob(jobType string, payload map[string]interface{}, maxRetries int) *secondary.Job {
	now := time.Now()
	return &secondary.Job{
		ID:         GenerateJobID(),
		Type:       jobType,
		Payload:    payload,
		Status:     secondary.JobStatusPending,
		Retries:    0,
		MaxRetries: maxRetries,
		CreatedAt:  now,
	}
}

// SerializeJob serializes a job to JSON
func SerializeJob(job *secondary.Job) ([]byte, error) {
	return json.Marshal(job)
}

// DeserializeJob deserializes a job from JSON
func DeserializeJob(data []byte) (*secondary.Job, error) {
	var job secondary.Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}
	return &job, nil
}

// ShouldRetry checks if a job should be retried based on retry count and max retries
func ShouldRetry(job *secondary.Job) bool {
	return job.Retries < job.MaxRetries
}

// IncrementRetry increments the retry count of a job
func IncrementRetry(job *secondary.Job) {
	job.Retries++
}

// CalculateRetryDelay calculates the delay before retrying a job (exponential backoff)
func CalculateRetryDelay(baseDelaySeconds int, retryCount int) time.Duration {
	// Exponential backoff: baseDelay * 2^retryCount
	// Max delay: 1 hour
	maxDelay := 1 * time.Hour
	baseDelay := time.Duration(baseDelaySeconds) * time.Second
	
	delay := baseDelay
	for i := 0; i < retryCount; i++ {
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
			break
		}
	}
	
	return delay
}

