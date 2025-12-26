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

// Ensure RedisRepository implements ICacheRepository
var _ secondary.ICacheRepository = (*RedisRepository)(nil)
