package secondary

import (
	"context"
	"time"
)

// ICacheRepository defines the interface for cache operations
// Implementation agnostic - works with Redis, Memcached, etc.
type ICacheRepository interface {
	// Ping verifies the cache connection
	Ping(ctx context.Context) error

	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// SetNX sets a key only if it does not exist (atomic operation)
	// Returns true if key was set, false if key already exists
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)

	// Expire sets the expiration time for a key
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// TTL returns the remaining time to live of a key
	// Returns -1 if key exists but has no expiration
	// Returns -2 if key does not exist
	TTL(ctx context.Context, key string) (time.Duration, error)
}

