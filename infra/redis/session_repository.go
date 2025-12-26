package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

const defaultSessionKeyPrefix = "session"

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

// Ensure SessionRepository implements ISessionRepository
var _ secondary.ISessionRepository = (*SessionRepository)(nil)

