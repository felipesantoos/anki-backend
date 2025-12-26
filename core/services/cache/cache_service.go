package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
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
	return s.repo.Get(ctx, key)
}

// Set stores a value in cache with TTL
func (s *CacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.repo.Set(ctx, key, value, ttl)
}

// Delete removes a key from cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.repo.Delete(ctx, key)
}

// GetOrSet implements cache-aside pattern
// If key exists, returns cached value
// If key doesn't exist, calls fetchFunc to get value, stores it, and returns it
func (s *CacheService) GetOrSet(ctx context.Context, key string, fetchFunc func() (string, error), ttl time.Duration) (string, error) {
	// Try to get from cache
	value, err := s.repo.Get(ctx, key)
	if err == nil {
		// Cache hit - return cached value
		return value, nil
	}

	// Check if error is "key not found" (cache miss) or another error
	// If it's another error (e.g., connection error), we should probably return it
	// For now, we'll treat any error as cache miss and continue
	// In production, you might want to log non-"not found" errors

	// Cache miss - fetch from source
	fetchedValue, err := fetchFunc()
	if err != nil {
		return "", fmt.Errorf("failed to fetch value: %w", err)
	}

	// Store in cache (best effort - don't fail if cache set fails)
	if err := s.repo.Set(ctx, key, fetchedValue, ttl); err != nil {
		// Log error but don't fail the operation
		// Return the fetched value even if cache set failed
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
	return s.repo.Exists(ctx, key)
}

