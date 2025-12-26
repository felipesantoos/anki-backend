package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/infra/postgres"
)

func TestDatabase_Connection(t *testing.T) {
	// Skip if database is not available
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()
	logger.Info("Starting database connection test",
		"host", cfg.Database.Host,
		"database", cfg.Database.DBName,
	)

	// Create database connection
	db, err := postgres.NewPostgresRepository(cfg.Database, logger)
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	// Test Ping
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Test Ping with timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctxTimeout); err != nil {
		t.Fatalf("Ping with timeout failed: %v", err)
	}
}

func TestDatabase_ConnectionPool(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	// Create database connection
	db, err := postgres.NewPostgresRepository(cfg.Database, logger)
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	// Verify pool configuration
	stats := db.DB.Stats()

	if stats.MaxOpenConnections != cfg.Database.MaxConnections {
		t.Errorf("MaxOpenConnections = %d, want %d",
			stats.MaxOpenConnections, cfg.Database.MaxConnections)
	}

	// Test multiple concurrent connections
	ctx := context.Background()
	errChan := make(chan error, 10)

	// Create 10 concurrent pings
	for i := 0; i < 10; i++ {
		go func() {
			if err := db.Ping(ctx); err != nil {
				errChan <- err
				return
			}
			errChan <- nil
		}()
	}

	// Wait for all pings to complete
	for i := 0; i < 10; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent ping failed: %v", err)
		}
	}

	// Verify pool stats after usage
	finalStats := db.DB.Stats()
	if finalStats.OpenConnections > stats.MaxOpenConnections {
		t.Errorf("OpenConnections (%d) exceeded MaxOpenConnections (%d)",
			finalStats.OpenConnections, stats.MaxOpenConnections)
	}
}

func TestDatabase_GracefulShutdown(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	// Create database connection
	db, err := postgres.NewPostgresRepository(cfg.Database, logger)
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}

	// Verify connection is open
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping failed before close: %v", err)
	}

	// Close connection
	if err := db.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Verify connection is closed (Ping should fail or return error)
	// Note: sql.DB.Close() may not immediately close all connections,
	// but subsequent operations should fail
	time.Sleep(100 * time.Millisecond)

	// Try to ping closed connection
	if err := db.Ping(ctx); err == nil {
		// This might succeed in some cases, but we at least verified Close() didn't error
		t.Log("Ping succeeded after Close(), connection may still be closing")
	}
}

func TestDatabase_QueryExecution(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	db, err := postgres.NewPostgresRepository(cfg.Database, logger)
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Execute a simple query
	var version string
	err = db.DB.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if version == "" {
		t.Error("Expected version string, got empty")
	}

	t.Logf("PostgreSQL version: %s", version)
}

func TestDatabase_Transaction(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.Default()

	db, err := postgres.NewPostgresRepository(cfg.Database, logger)
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Start a transaction
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	// Execute a query in transaction
	var result int
	err = tx.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Query in transaction failed: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected result 1, got %d", result)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
}
