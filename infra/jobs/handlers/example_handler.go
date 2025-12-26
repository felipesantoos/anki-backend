package handlers

import (
	"context"
	"fmt"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// ExampleHandler is an example job handler implementation
// This can be removed or replaced with actual job handlers
type ExampleHandler struct {
	jobType string
}

// NewExampleHandler creates a new example handler
func NewExampleHandler(jobType string) *ExampleHandler {
	return &ExampleHandler{
		jobType: jobType,
	}
}

// Handle processes the example job
func (h *ExampleHandler) Handle(ctx context.Context, job *secondary.Job) error {
	// Example: log the job payload
	fmt.Printf("Processing example job: %s with payload: %+v\n", job.ID, job.Payload)
	
	// Simulate some work
	// In a real handler, you would perform actual operations here
	
	return nil
}

// JobType returns the type of job this handler processes
func (h *ExampleHandler) JobType() string {
	return h.jobType
}

