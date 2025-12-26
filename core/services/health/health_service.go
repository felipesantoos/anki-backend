package health

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// HealthService implements IHealthService for health check operations
// This service follows hexagonal architecture principles:
// - Uses interfaces (IDatabaseRepository, ICacheRepository) instead of concrete implementations
// - Implementation-agnostic: doesn't know about PostgreSQL, Redis, MySQL, Memcached, etc.
// - Testable: can be easily mocked for unit tests
type HealthService struct {
	db    secondary.IDatabaseRepository
	cache secondary.ICacheRepository
}

// NewHealthService creates a new HealthService instance
func NewHealthService(db secondary.IDatabaseRepository, cache secondary.ICacheRepository) primary.IHealthService {
	return &HealthService{
		db:    db,
		cache: cache,
	}
}

// CheckHealth performs health checks on database and cache, returning a health response
func (s *HealthService) CheckHealth(ctx context.Context) (*response.HealthResponse, error) {
	healthResp := response.NewHealthResponse()

	// Create context with timeout for health checks
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check database health
	s.checkDatabase(healthCtx, healthResp)

	// Check cache health
	s.checkCache(healthCtx, healthResp)

	// Calculate overall status
	healthResp.CalculateOverallStatus()

	return healthResp, nil
}

// checkDatabase checks the database connection health
func (s *HealthService) checkDatabase(ctx context.Context, healthResp *response.HealthResponse) {
	if s.db == nil {
		healthResp.SetComponent("database", "unhealthy", "Database connection not initialized")
		return
	}

	if err := s.db.Ping(ctx); err != nil {
		healthResp.SetComponent("database", "unhealthy", err.Error())
		return
	}

	healthResp.SetComponent("database", "healthy", "Connection successful")
}

// checkCache checks the cache connection health
func (s *HealthService) checkCache(ctx context.Context, healthResp *response.HealthResponse) {
	if s.cache == nil {
		healthResp.SetComponent("redis", "unhealthy", "Cache connection not initialized")
		return
	}

	if err := s.cache.Ping(ctx); err != nil {
		healthResp.SetComponent("redis", "unhealthy", err.Error())
		return
	}

	healthResp.SetComponent("redis", "healthy", "Connection successful")
}

