package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

const (
	defaultSessionKeyPrefix        = "session"
	userSessionsKeyPrefix          = "user_sessions"
	refreshTokenSessionKeyPrefix   = "refresh_token_session"
)

// SessionRepository implements ISessionRepository using Redis
type SessionRepository struct {
	client *redis.Client
	prefix string
}

// NewSessionRepository creates a new SessionRepository instance
func NewSessionRepository(client *redis.Client, keyPrefix string) secondary.ISessionRepository {
	prefix := keyPrefix
	if prefix == "" {
		prefix = defaultSessionKeyPrefix
	}

	return &SessionRepository{
		client: client,
		prefix: prefix,
	}
}

// buildKey builds the Redis key for a sessionID
func (r *SessionRepository) buildKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", r.prefix, sessionID)
}

// buildUserSessionsKey builds the Redis key for a user's session set
func (r *SessionRepository) buildUserSessionsKey(userID int64) string {
	return fmt.Sprintf("%s:%d", userSessionsKeyPrefix, userID)
}

// buildRefreshTokenSessionKey builds the Redis key for refresh token to session mapping
func (r *SessionRepository) buildRefreshTokenSessionKey(refreshTokenHash string) string {
	return fmt.Sprintf("%s:%s", refreshTokenSessionKeyPrefix, refreshTokenHash)
}

// GetSession retrieves session data by sessionID
func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	key := r.buildKey(sessionID)

	// Get value from Redis
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Deserialize JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return data, nil
}

// SetSession stores session data with TTL
func (r *SessionRepository) SetSession(ctx context.Context, sessionID string, data map[string]interface{}, ttl time.Duration) error {
	key := r.buildKey(sessionID)

	// Serialize to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Store in Redis with TTL
	if err := r.client.Set(ctx, key, string(jsonData), ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	return nil
}

// DeleteSession removes a session
func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	key := r.buildKey(sessionID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// RefreshSession updates the TTL of an existing session
func (r *SessionRepository) RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := r.buildKey(sessionID)

	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	return nil
}

// Exists checks if a session exists
func (r *SessionRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := r.buildKey(sessionID)

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// GetUserSessions retrieves all session IDs for a user
func (r *SessionRepository) GetUserSessions(ctx context.Context, userID int64) ([]string, error) {
	key := r.buildUserSessionsKey(userID)

	// Get all members of the set
	sessionIDs, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessionIDs, nil
}

// AddUserSession adds a session ID to the user's session set
func (r *SessionRepository) AddUserSession(ctx context.Context, userID int64, sessionID string) error {
	key := r.buildUserSessionsKey(userID)

	// Add session ID to set
	if err := r.client.SAdd(ctx, key, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to add user session: %w", err)
	}

	// Set TTL on the set (use a longer TTL than individual sessions to allow cleanup)
	// Using 90 days as default, but this should ideally match refresh token expiry
	if err := r.client.Expire(ctx, key, 90*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set TTL on user sessions set: %w", err)
	}

	return nil
}

// RemoveUserSession removes a session ID from the user's session set
func (r *SessionRepository) RemoveUserSession(ctx context.Context, userID int64, sessionID string) error {
	key := r.buildUserSessionsKey(userID)

	// Remove session ID from set
	if err := r.client.SRem(ctx, key, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to remove user session: %w", err)
	}

	return nil
}

// GetSessionByRefreshToken retrieves session ID associated with a refresh token hash
func (r *SessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (string, error) {
	key := r.buildRefreshTokenSessionKey(refreshTokenHash)

	// Get session ID from Redis
	sessionID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("refresh token session not found: %s", refreshTokenHash)
		}
		return "", fmt.Errorf("failed to get refresh token session: %w", err)
	}

	return sessionID, nil
}

// SetRefreshTokenSession associates a refresh token hash with a session ID
func (r *SessionRepository) SetRefreshTokenSession(ctx context.Context, refreshTokenHash string, sessionID string, ttl time.Duration) error {
	key := r.buildRefreshTokenSessionKey(refreshTokenHash)

	// Store session ID with TTL
	if err := r.client.Set(ctx, key, sessionID, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set refresh token session: %w", err)
	}

	return nil
}

// DeleteRefreshTokenSession removes the association between refresh token and session
func (r *SessionRepository) DeleteRefreshTokenSession(ctx context.Context, refreshTokenHash string) error {
	key := r.buildRefreshTokenSessionKey(refreshTokenHash)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token session: %w", err)
	}

	return nil
}

// Ensure SessionRepository implements ISessionRepository
var _ secondary.ISessionRepository = (*SessionRepository)(nil)

