package health

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
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
	// Create span for health check operation
	ctx, span := tracing.StartSpan(ctx, "health.check")
	defer span.End()

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

	// Add overall status to span
	span.SetAttributes(
		attribute.String("health.status", healthResp.Status),
	)

	return healthResp, nil
}

// checkDatabase checks the database connection health
func (s *HealthService) checkDatabase(ctx context.Context, healthResp *response.HealthResponse) {
	ctx, span := tracing.StartSpan(ctx, "health.check.database")
	defer span.End()

	if s.db == nil {
		healthResp.SetComponent("database", "unhealthy", "Database connection not initialized")
		span.SetAttributes(attribute.String("database.status", "unhealthy"))
		return
	}

	if err := s.db.Ping(ctx); err != nil {
		healthResp.SetComponent("database", "unhealthy", err.Error())
		tracing.RecordError(span, err)
		span.SetAttributes(attribute.String("database.status", "unhealthy"))
		return
	}

	healthResp.SetComponent("database", "healthy", "Connection successful")
	span.SetAttributes(attribute.String("database.status", "healthy"))
}

// checkCache checks the cache connection health
func (s *HealthService) checkCache(ctx context.Context, healthResp *response.HealthResponse) {
	ctx, span := tracing.StartSpan(ctx, "health.check.cache")
	defer span.End()

	if s.cache == nil {
		healthResp.SetComponent("redis", "unhealthy", "Cache connection not initialized")
		span.SetAttributes(attribute.String("cache.status", "unhealthy"))
		return
	}

	if err := s.cache.Ping(ctx); err != nil {
		healthResp.SetComponent("redis", "unhealthy", err.Error())
		tracing.RecordError(span, err)
		span.SetAttributes(attribute.String("cache.status", "unhealthy"))
		return
	}

	healthResp.SetComponent("redis", "healthy", "Connection successful")
	span.SetAttributes(attribute.String("cache.status", "healthy"))
}

