package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CacheService provides high-level cache operations
// Uses cache repository interface (Redis, Memcached, etc.)
type CacheService struct {
	repo secondary.ICacheRepository
}

// NewCacheService creates a new CacheService instance
func NewCacheService(repo secondary.ICacheRepository) *CacheService {
	return &CacheService{
		repo: repo,
	}
}

// Get retrieves a value from cache by key
func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "cache.get",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	result, err := s.repo.Get(ctx, key)
	if err != nil {
		tracing.RecordError(span, err)
		return "", err
	}
	span.SetAttributes(attribute.Bool("cache.hit", true))
	return result, nil
}

// Set stores a value in cache with TTL
func (s *CacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	ctx, span := tracing.StartSpan(ctx, "cache.set",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.ttl", ttl.String()),
		),
	)
	defer span.End()

	err := s.repo.Set(ctx, key, value, ttl)
	if err != nil {
		tracing.RecordError(span, err)
		return err
	}
	return nil
}

// Delete removes a key from cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	ctx, span := tracing.StartSpan(ctx, "cache.delete",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	err := s.repo.Delete(ctx, key)
	if err != nil {
		tracing.RecordError(span, err)
		return err
	}
	return nil
}

// GetOrSet implements cache-aside pattern
// If key exists, returns cached value
// If key doesn't exist, calls fetchFunc to get value, stores it, and returns it
func (s *CacheService) GetOrSet(ctx context.Context, key string, fetchFunc func() (string, error), ttl time.Duration) (string, error) {
	ctx, span := tracing.StartSpan(ctx, "cache.get_or_set",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.ttl", ttl.String()),
		),
	)
	defer span.End()

	// Try to get from cache
	value, err := s.repo.Get(ctx, key)
	if err == nil {
		// Cache hit - return cached value
		span.SetAttributes(attribute.Bool("cache.hit", true))
		return value, nil
	}

	// Cache miss - fetch from source
	span.SetAttributes(attribute.Bool("cache.hit", false))
	fetchedValue, err := fetchFunc()
	if err != nil {
		tracing.RecordError(span, err)
		return "", fmt.Errorf("failed to fetch value: %w", err)
	}

	// Store in cache (best effort - don't fail if cache set fails)
	if err := s.repo.Set(ctx, key, fetchedValue, ttl); err != nil {
		// Log error but don't fail the operation
		// Return the fetched value even if cache set failed
		tracing.RecordError(span, err)
	}

	return fetchedValue, nil
}

// InvalidatePattern deletes all keys matching a pattern
// Note: This requires SCAN operation which may be slow on large datasets
func (s *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	// Pattern invalidation requires repository-specific implementation
	// For now, return an error indicating this needs to be implemented
	// In a real implementation, this would use Redis SCAN or similar
	return fmt.Errorf("pattern invalidation not implemented - use specific key deletion or implement SCAN in repository")
}

// Exists checks if a key exists in cache
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "cache.exists",
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	exists, err := s.repo.Exists(ctx, key)
	if err != nil {
		tracing.RecordError(span, err)
		return false, err
	}
	span.SetAttributes(attribute.Bool("cache.exists", exists))
	return exists, nil
}

