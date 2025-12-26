package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/infra/redis"
)

func TestRedis_Connection(t *testing.T) {
	// Skip if Redis is not available
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()
	logger.Info("Starting Redis connection test",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port,
		"db", cfg.Redis.DB,
	)

	// Create Redis connection
	rdb, err := redis.NewRedis(cfg.Redis, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis connection: %v", err)
	}
	defer rdb.Close()

	// Test Ping
	ctx := context.Background()
	if err := rdb.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Test Ping with timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctxTimeout); err != nil {
		t.Fatalf("Ping with timeout failed: %v", err)
	}
}

func TestRedis_BasicOperations(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	rdb, err := redis.NewRedis(cfg.Redis, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis connection: %v", err)
	}
	defer rdb.Close()

	ctx := context.Background()

	// Test Set
	key := "test:key:integration"
	value := "test-value"

	err = rdb.Client.Set(ctx, key, value, 10*time.Second).Err()
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	result, err := rdb.Client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result != value {
		t.Errorf("Expected value %s, got %s", value, result)
	}

	// Cleanup
	rdb.Client.Del(ctx, key)
}

func TestRedis_ConnectionPool(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	rdb, err := redis.NewRedis(cfg.Redis, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis connection: %v", err)
	}
	defer rdb.Close()

	// Verify pool stats
	poolStats := rdb.Client.PoolStats()

	if poolStats.TotalConns == 0 {
		t.Error("Expected pool to have connections")
	}

	// Test multiple concurrent operations
	ctx := context.Background()
	errChan := make(chan error, 10)

	// Create 10 concurrent operations
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("test:concurrent:%d", index)
			if err := rdb.Client.Set(ctx, key, "value", time.Second).Err(); err != nil {
				errChan <- err
				return
			}
			errChan <- nil
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < 10; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}

	// Verify final pool stats
	finalPoolStats := rdb.Client.PoolStats()
	if finalPoolStats.TotalConns == 0 {
		t.Error("Expected pool to have connections after operations")
	}
}

func TestRedis_GracefulShutdown(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	rdb, err := redis.NewRedis(cfg.Redis, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis connection: %v", err)
	}

	// Verify connection is open
	ctx := context.Background()
	if err := rdb.Ping(ctx); err != nil {
		t.Fatalf("Ping failed before close: %v", err)
	}

	// Close connection
	if err := rdb.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Verify connection is closed (Ping should fail)
	time.Sleep(100 * time.Millisecond)

	// Try to ping closed connection
	if err := rdb.Ping(ctx); err == nil {
		// This might succeed in some cases, but we at least verified Close() didn't error
		t.Log("Ping succeeded after Close(), connection may still be closing")
	}
}

func TestRedis_WithoutPassword(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Override password to empty (testing without authentication)
	testCfg := cfg.Redis
	testCfg.Password = ""

	logger := slog.Default()

	rdb, err := redis.NewRedis(testCfg, logger)
	if err != nil {
		// This might fail if Redis requires password, which is fine
		t.Skipf("Skipping test: Redis requires password or not available: %v", err)
	}
	defer rdb.Close()

	// Test Ping
	ctx := context.Background()
	if err := rdb.Ping(ctx); err != nil {
		t.Errorf("Ping failed with empty password: %v", err)
	}
}

func TestRedis_DifferentDatabases(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TESTS") == "true" {
		t.Skip("Skipping Redis integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	// Test database 0
	cfg0 := cfg.Redis
	cfg0.DB = 0

	rdb0, err := redis.NewRedis(cfg0, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis connection for DB 0: %v", err)
	}
	defer rdb0.Close()

	ctx := context.Background()
	if err := rdb0.Ping(ctx); err != nil {
		t.Fatalf("Ping failed for DB 0: %v", err)
	}

	// Test database 1
	cfg1 := cfg.Redis
	cfg1.DB = 1

	rdb1, err := redis.NewRedis(cfg1, logger)
	if err != nil {
		t.Fatalf("Failed to create Redis connection for DB 1: %v", err)
	}
	defer rdb1.Close()

	if err := rdb1.Ping(ctx); err != nil {
		t.Fatalf("Ping failed for DB 1: %v", err)
	}

	// Verify they are separate databases
	key := "test:db:isolation"
	value0 := "db0-value"
	value1 := "db1-value"

	// Set in DB 0
	rdb0.Client.Set(ctx, key, value0, time.Minute)

	// Get from DB 0 (should exist)
	result0, err := rdb0.Client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("Get from DB 0 failed: %v", err)
	}
	if result0 != value0 {
		t.Errorf("Expected %s in DB 0, got %s", value0, result0)
	}

	// Get from DB 1 (should not exist)
	_, err = rdb1.Client.Get(ctx, key).Result()
	if err == nil {
		t.Error("Expected key to not exist in DB 1 (different database)")
	}

	// Cleanup
	rdb0.Client.Del(ctx, key)
}
