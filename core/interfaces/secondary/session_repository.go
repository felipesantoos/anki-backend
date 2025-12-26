package secondary

import (
	"context"
	"time"
)

// ISessionRepository defines the interface for session management operations
// Implementation agnostic - works with Redis, Memcached, database, etc.
type ISessionRepository interface {
	// GetSession retrieves session data by sessionID
	GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error)

	// SetSession stores session data with TTL
	SetSession(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error

	// DeleteSession removes a session
	DeleteSession(ctx context.Context, sessionID string) error

	// RefreshSession updates the TTL of an existing session
	RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error

	// Exists checks if a session exists
	Exists(ctx context.Context, sessionID string) (bool, error)
}

