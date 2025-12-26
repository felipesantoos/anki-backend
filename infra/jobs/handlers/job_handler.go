package handlers

import (
	"context"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// JobHandler defines the interface for processing a specific type of job
type JobHandler interface {
	// Handle processes a job
	Handle(ctx context.Context, job *secondary.Job) error

	// JobType returns the type of job this handler processes
	JobType() string
}

