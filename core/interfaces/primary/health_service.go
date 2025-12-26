package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
)

// IHealthService defines the interface for health check operations
type IHealthService interface {
	CheckHealth(ctx context.Context) (*response.HealthResponse, error)
}

