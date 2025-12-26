package redis

import (
	"log/slog"
	"testing"

	"github.com/felipesantos/anki-backend/config"
)

// TestNewRedisRepository_InvalidHost tests that NewRedisRepository returns an error for invalid host
// This is a unit test that doesn't require a real Redis connection
func TestNewRedisRepository_InvalidHost(t *testing.T) {
	logger := slog.Default()

	cfg := config.RedisConfig{
		Host:     "invalid-host",
		Port:     "6379",
		Password: "",
		DB:       0,
	}

	// This should fail to connect
	rdb, err := NewRedisRepository(cfg, logger)
	if err == nil {
		// If it somehow connects, close it
		if rdb != nil {
			rdb.Close()
		}
		t.Error("Expected error for invalid Redis host, got nil")
	}
}

// Note: Tests that require real Redis connections (Ping, Close, PoolConfiguration, etc.)
// are located in tests/integration/redis_test.go to avoid duplication and
// keep unit tests in infra/ focused on unit-level testing without external dependencies.
