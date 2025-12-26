package redis

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/felipesantos/anki-backend/config"
)

// Redis wraps the redis.Client with additional functionality
type Redis struct {
	Client *redis.Client
}

// NewRedis creates a new Redis connection with connection pooling configured
func NewRedis(cfg config.RedisConfig, logger *slog.Logger) (*Redis, error) {
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

	return &Redis{Client: client}, nil
}

// Ping verifies the Redis connection
func (r *Redis) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Close closes the Redis connection gracefully
func (r *Redis) Close() error {
	return r.Client.Close()
}
