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

	// GetUserSessions retrieves all session IDs for a user
	GetUserSessions(ctx context.Context, userID int64) ([]string, error)

	// AddUserSession adds a session ID to the user's session set
	AddUserSession(ctx context.Context, userID int64, sessionID string) error

	// RemoveUserSession removes a session ID from the user's session set
	RemoveUserSession(ctx context.Context, userID int64, sessionID string) error

	// GetSessionByRefreshToken retrieves session ID associated with a refresh token hash
	GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (string, error)

	// SetRefreshTokenSession associates a refresh token hash with a session ID
	SetRefreshTokenSession(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error

	// DeleteRefreshTokenSession removes the association between refresh token and session
	DeleteRefreshTokenSession(ctx context.Context, refreshTokenHash string) error
}

