package redis

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
)

func TestNewRedis_WithPassword(t *testing.T) {
	// Skip if Redis is not available
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis tests")
	}

	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "testpassword",
		DB:       0,
	}

	// This test requires a real Redis connection
	// Skip if Redis is not available
	rdb, err := NewRedis(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer rdb.Close()

	if rdb.Client == nil {
		t.Error("Expected Redis client, got nil")
	}
}

func TestNewRedis_WithoutPassword(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis tests")
	}

	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "", // No password
		DB:       0,
	}

	rdb, err := NewRedis(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer rdb.Close()

	if rdb.Client == nil {
		t.Error("Expected Redis client, got nil")
	}
}

func TestNewRedis_DifferentDB(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis tests")
	}

	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       1, // Different database
	}

	rdb, err := NewRedis(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer rdb.Close()

	if rdb.Client == nil {
		t.Error("Expected Redis client, got nil")
	}
}

func TestRedis_Ping(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis tests")
	}

	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       0,
	}

	rdb, err := NewRedis(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test with timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := rdb.Ping(ctxTimeout); err != nil {
		t.Errorf("Ping() with timeout error = %v", err)
	}
}

func TestRedis_Close(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis tests")
	}

	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       0,
	}

	rdb, err := NewRedis(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}

	// Close should not return error
	if err := rdb.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Close again should not panic (idempotent)
	if err := rdb.Close(); err != nil {
		// It's okay if it returns an error on second close
		_ = err
	}
}

func TestNewRedis_InvalidHost(t *testing.T) {
	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "invalid-host",
		Port:     "6379",
		Password: "",
		DB:       0,
	}

	// This should fail to connect
	rdb, err := NewRedis(cfg, logger)
	if err == nil {
		// If it somehow connects, close it
		if rdb != nil {
			rdb.Close()
		}
		t.Error("Expected error for invalid Redis host, got nil")
	}
}

func TestRedis_PoolConfiguration(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis tests")
	}

	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       0,
	}

	rdb, err := NewRedis(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: Redis not available: %v", err)
	}
	defer rdb.Close()

	// Verify pool stats
	poolStats := rdb.Client.PoolStats()

	// PoolSize should be 10 (default)
	if poolStats.TotalConns == 0 {
		t.Error("Expected pool to have connections")
	}

	// Test Ping to verify connection works
	ctx := context.Background()
	if err := rdb.Ping(ctx); err != nil {
		t.Errorf("Ping() after pool config error = %v", err)
	}
}
