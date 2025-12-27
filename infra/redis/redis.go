package redis

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

var tracer = otel.Tracer("anki-backend/redis")

// RedisRepository wraps the redis.Client with additional functionality
// Implements ICacheRepository interface
type RedisRepository struct {
	Client *redis.Client
	db     int
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

	// Create Redis client
	// OpenTelemetry instrumentation is implemented manually using spans
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

	return &RedisRepository{
		Client: client,
		db:     cfg.DB,
	}, nil
}

// Ping verifies the Redis connection
// Implements ICacheRepository interface
func (r *RedisRepository) Ping(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "redis.ping",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.Int("db.redis.database_index", r.db),
		),
	)
	defer span.End()

	err := r.Client.Ping(ctx).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// Close closes the Redis connection gracefully
func (r *RedisRepository) Close() error {
	return r.Client.Close()
}

// Get retrieves a value from cache by key
// Returns an error if the key does not exist
func (r *RedisRepository) Get(ctx context.Context, key string) (string, error) {
	ctx, span := tracer.Start(ctx, "redis.get",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "get"),
			attribute.String("db.redis.key", key),
			attribute.Int("db.redis.database_index", r.db),
		),
	)
	defer span.End()

	result, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			span.SetStatus(codes.Ok, "key not found")
			return "", fmt.Errorf("key not found: %s", key)
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

// Set stores a value in cache with TTL
func (r *RedisRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	ctx, span := tracer.Start(ctx, "redis.set",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "set"),
			attribute.String("db.redis.key", key),
			attribute.Int("db.redis.database_index", r.db),
			attribute.String("db.redis.command.ttl", ttl.String()),
		),
	)
	defer span.End()

	err := r.Client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// Delete removes a key from cache
func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	ctx, span := tracer.Start(ctx, "redis.delete",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "delete"),
			attribute.String("db.redis.key", key),
			attribute.Int("db.redis.database_index", r.db),
		),
	)
	defer span.End()

	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// Exists checks if a key exists in cache
func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	ctx, span := tracer.Start(ctx, "redis.exists",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "exists"),
			attribute.String("db.redis.key", key),
			attribute.Int("db.redis.database_index", r.db),
		),
	)
	defer span.End()

	count, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}
	span.SetStatus(codes.Ok, "")
	return count > 0, nil
}

// SetNX sets a key only if it does not exist (atomic operation)
// Returns true if key was set, false if key already exists
func (r *RedisRepository) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	ctx, span := tracer.Start(ctx, "redis.setnx",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "setnx"),
			attribute.String("db.redis.key", key),
			attribute.Int("db.redis.database_index", r.db),
			attribute.String("db.redis.command.ttl", ttl.String()),
		),
	)
	defer span.End()

	result, err := r.Client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

// Expire sets the expiration time for a key
func (r *RedisRepository) Expire(ctx context.Context, key string, ttl time.Duration) error {
	ctx, span := tracer.Start(ctx, "redis.expire",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "expire"),
			attribute.String("db.redis.key", key),
			attribute.Int("db.redis.database_index", r.db),
			attribute.String("db.redis.command.ttl", ttl.String()),
		),
	)
	defer span.End()

	err := r.Client.Expire(ctx, key, ttl).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}

// TTL returns the remaining time to live of a key
// Returns -1 if key exists but has no expiration
// Returns -2 if key does not exist
func (r *RedisRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	ctx, span := tracer.Start(ctx, "redis.ttl",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "ttl"),
			attribute.String("db.redis.key", key),
			attribute.Int("db.redis.database_index", r.db),
		),
	)
	defer span.End()

	result, err := r.Client.TTL(ctx, key).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	span.SetStatus(codes.Ok, "")
	return result, nil
}

// Ensure RedisRepository implements ICacheRepository
var _ secondary.ICacheRepository = (*RedisRepository)(nil)
