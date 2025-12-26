package redis

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// RedisRepository wraps the redis.Client with additional functionality
// Implements ICacheRepository interface
type RedisRepository struct {
	Client *redis.Client
}

// NewRedisRepository creates a new Redis connection with connection pooling configured
func NewRedisRepository(cfg config.RedisConfig, logger *slog.Logger) (*RedisRepository, error) {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	logger.Info("Connecting to Redis",
		"host", cfg.Host,
		"port", cfg.Port,
		"db", cfg.DB,
	)

	options := &redis.Options{
		Addr:         addr,
		DB:           cfg.DB,
		PoolSize:     10,                      // Default pool size
		MinIdleConns: 5,                       // Minimum idle connections
		MaxRetries:   3,                       // Maximum retries
		DialTimeout:  5 * time.Second,         // Timeout for connecting
		ReadTimeout:  3 * time.Second,         // Timeout for reading
		WriteTimeout: 3 * time.Second,         // Timeout for writing
	}

	// Only set password if provided (support Redis without authentication)
	if cfg.Password != "" {
		options.Password = cfg.Password
	}

	client := redis.NewClient(options)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connection established",
		"host", cfg.Host,
		"port", cfg.Port,
		"db", cfg.DB,
		"pool_size", options.PoolSize,
		"min_idle_conns", options.MinIdleConns,
	)

	return &RedisRepository{Client: client}, nil
}

// Ping verifies the Redis connection
// Implements ICacheRepository interface
func (r *RedisRepository) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Close closes the Redis connection gracefully
func (r *RedisRepository) Close() error {
	return r.Client.Close()
}

// Get retrieves a value from cache by key
// Returns an error if the key does not exist
func (r *RedisRepository) Get(ctx context.Context, key string) (string, error) {
	result, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", err
	}
	return result, nil
}

// Set stores a value in cache with TTL
func (r *RedisRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from cache
func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache
func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SetNX sets a key only if it does not exist (atomic operation)
// Returns true if key was set, false if key already exists
func (r *RedisRepository) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	result, err := r.Client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// Expire sets the expiration time for a key
func (r *RedisRepository) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.Client.Expire(ctx, key, ttl).Err()
}

// TTL returns the remaining time to live of a key
// Returns -1 if key exists but has no expiration
// Returns -2 if key does not exist
func (r *RedisRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := r.Client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Ensure RedisRepository implements ICacheRepository
var _ secondary.ICacheRepository = (*RedisRepository)(nil)
