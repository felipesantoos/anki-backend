package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/core/services/jobs"
)

// mockJobQueue is a mock implementation of IJobQueue for testing
type mockJobQueue struct {
	enqueueFunc    func(ctx context.Context, job *secondary.Job) error
	dequeueFunc    func(ctx context.Context, timeout time.Duration) (*secondary.Job, error)
	getStatusFunc  func(ctx context.Context, jobID string) (*secondary.Job, error)
	updateStatusFunc func(ctx context.Context, jobID string, status secondary.JobStatus) error
	retryFunc      func(ctx context.Context, job *secondary.Job) error
}

func (m *mockJobQueue) Enqueue(ctx context.Context, job *secondary.Job) error {
	if m.enqueueFunc != nil {
		return m.enqueueFunc(ctx, job)
	}
	return nil
}

func (m *mockJobQueue) Dequeue(ctx context.Context, timeout time.Duration) (*secondary.Job, error) {
	if m.dequeueFunc != nil {
		return m.dequeueFunc(ctx, timeout)
	}
	return nil, errors.New("not implemented")
}

func (m *mockJobQueue) GetStatus(ctx context.Context, jobID string) (*secondary.Job, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx, jobID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockJobQueue) UpdateStatus(ctx context.Context, jobID string, status secondary.JobStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, jobID, status)
	}
	return errors.New("not implemented")
}

func (m *mockJobQueue) Retry(ctx context.Context, job *secondary.Job) error {
	if m.retryFunc != nil {
		return m.retryFunc(ctx, job)
	}
	return errors.New("not implemented")
}

func TestJobService_Enqueue(t *testing.T) {
	tests := []struct {
		name       string
		jobType    string
		payload    map[string]interface{}
		maxRetries int
		wantErr    bool
	}{
		{
			name:       "successful enqueue",
			jobType:    "test_job",
			payload:    map[string]interface{}{"key": "value"},
			maxRetries: 3,
			wantErr:    false,
		},
		{
			name:       "empty job type",
			jobType:    "",
			payload:    map[string]interface{}{},
			maxRetries: 3,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQueue := &mockJobQueue{
				enqueueFunc: func(ctx context.Context, job *secondary.Job) error {
					if job.Type != tt.jobType {
						t.Errorf("expected job type %s, got %s", tt.jobType, job.Type)
					}
					if job.MaxRetries != tt.maxRetries {
						t.Errorf("expected max retries %d, got %d", tt.maxRetries, job.MaxRetries)
					}
					return nil
				},
			}
			service := jobs.NewJobService(mockQueue, tt.maxRetries)

			ctx := context.Background()
			jobID, err := service.Enqueue(ctx, tt.jobType, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if jobID != "" {
					t.Errorf("expected empty job ID on error, got %s", jobID)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if jobID == "" {
					t.Errorf("expected non-empty job ID, got empty string")
				}
			}
		})
	}
}

func TestJobService_GetStatus(t *testing.T) {
	tests := []struct {
		name    string
		jobID   string
		wantErr bool
	}{
		{
			name:    "successful get status",
			jobID:   "test-job-id",
			wantErr: false,
		},
		{
			name:    "empty job ID",
			jobID:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQueue := &mockJobQueue{}
			if !tt.wantErr {
				mockQueue.getStatusFunc = func(ctx context.Context, jobID string) (*secondary.Job, error) {
					return &secondary.Job{
						ID:     jobID,
						Type:   "test_job",
						Status: secondary.JobStatusPending,
					}, nil
				}
			}
			service := jobs.NewJobService(mockQueue, 3)

			ctx := context.Background()
			job, err := service.GetStatus(ctx, tt.jobID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if job != nil {
					t.Errorf("expected nil job on error, got %+v", job)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if job == nil {
					t.Errorf("expected non-nil job, got nil")
				} else if job.ID != tt.jobID {
					t.Errorf("expected job ID %s, got %s", tt.jobID, job.ID)
				}
			}
		})
	}
}

func TestJobService_EnqueueWithRetries(t *testing.T) {
	mockQueue := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, job *secondary.Job) error {
			if job.Type != "test_job" {
				t.Errorf("expected job type test_job, got %s", job.Type)
			}
			if job.MaxRetries != 10 {
				t.Errorf("expected max retries 10, got %d", job.MaxRetries)
			}
			return nil
		},
	}
	service := jobs.NewJobService(mockQueue, 5) // Default max retries: 5

	payload := map[string]interface{}{"key": "value"}
	customMaxRetries := 10

	ctx := context.Background()
	jobID, err := service.EnqueueWithRetries(ctx, "test_job", payload, customMaxRetries)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if jobID == "" {
		t.Errorf("expected non-empty job ID, got empty string")
	}

	// Test with negative max retries
	_, err = service.EnqueueWithRetries(ctx, "test_job", payload, -1)
	if err == nil {
		t.Errorf("expected error for negative max retries, got nil")
	}
}

