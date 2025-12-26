package jobs

import (
	"fmt"
	"sync"

	"github.com/felipesantos/anki-backend/infra/jobs/handlers"
)

// JobRegistry maintains a mapping of job types to their handlers
type JobRegistry struct {
	mu       sync.RWMutex
	handlers map[string]handlers.JobHandler
}

// NewJobRegistry creates a new job registry
func NewJobRegistry() *JobRegistry {
	return &JobRegistry{
		handlers: make(map[string]handlers.JobHandler),
	}
}

// Register registers a handler for a specific job type
func (r *JobRegistry) Register(handler handlers.JobHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	jobType := handler.JobType()
	if jobType == "" {
		return fmt.Errorf("handler must return a non-empty job type")
	}

	if _, exists := r.handlers[jobType]; exists {
		return fmt.Errorf("handler for job type '%s' is already registered", jobType)
	}

	r.handlers[jobType] = handler
	return nil
}

// GetHandler retrieves a handler for a specific job type
func (r *JobRegistry) GetHandler(jobType string) (handlers.JobHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[jobType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for job type '%s'", jobType)
	}

	return handler, nil
}

// GetRegisteredTypes returns all registered job types
func (r *JobRegistry) GetRegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.handlers))
	for jobType := range r.handlers {
		types = append(types, jobType)
	}
	return types
}

