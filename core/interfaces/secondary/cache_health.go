package secondary

import "context"

// ICacheRepository defines the interface for cache health checks
// Implementation agnostic - works with Redis, Memcached, etc.
type ICacheRepository interface {
	Ping(ctx context.Context) error
}

