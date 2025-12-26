package database

import (
	"context"
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.DatabaseConfig
		expected string
	}{
		{
			name: "basic config",
			cfg: config.DatabaseConfig{
				Host:    "localhost",
				Port:    "5432",
				User:    "postgres",
				Password: "password",
				DBName:  "anki",
				SSLMode: "disable",
			},
			expected: "postgres://postgres:password@localhost:5432/anki?sslmode=disable",
		},
		{
			name: "password with special characters",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "p@ssw0rd!@#",
				DBName:   "anki",
				SSLMode:  "require",
			},
			expected: "postgres://postgres:" + url.QueryEscape("p@ssw0rd!@#") + "@localhost:5432/anki?sslmode=require",
		},
		{
			name: "production config",
			cfg: config.DatabaseConfig{
				Host:    "db.example.com",
				Port:    "5432",
				User:    "app_user",
				Password: "secure_password",
				DBName:  "anki_prod",
				SSLMode: "verify-full",
			},
			expected: "postgres://app_user:secure_password@db.example.com:5432/anki_prod?sslmode=verify-full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, err := buildDSN(tt.cfg)
			if err != nil {
				t.Fatalf("buildDSN() error = %v", err)
			}

			if dsn != tt.expected {
				t.Errorf("buildDSN() = %v, want %v", dsn, tt.expected)
			}
		})
	}
}

func TestNewDatabase_InvalidConfig(t *testing.T) {
	logger := slog.Default()

	cfg := config.DatabaseConfig{
		Host:     "invalid-host",
		Port:     "5432",
		User:     "user",
		Password: "password",
		DBName:   "database",
		SSLMode:  "disable",
	}

	// This should fail to connect, but not crash
	db, err := NewDatabase(cfg, logger)
	if err == nil {
		// If it somehow connects, close it
		if db != nil {
			db.Close()
		}
		t.Error("Expected error for invalid database config, got nil")
	}
}

func TestDatabase_Ping(t *testing.T) {
	logger := slog.Default()

	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		DBName:          "anki",
		SSLMode:         "disable",
		MaxConnections:  25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5,
	}

	// This test requires a real database connection
	// Skip if database is not available
	db, err := NewDatabase(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test with timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := db.Ping(ctxTimeout); err != nil {
		t.Errorf("Ping() with timeout error = %v", err)
	}
}

func TestDatabase_Close(t *testing.T) {
	logger := slog.Default()

	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		DBName:          "anki",
		SSLMode:         "disable",
		MaxConnections:  25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5,
	}

	// This test requires a real database connection
	db, err := NewDatabase(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
	}

	// Close should not return error
	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Close again should not panic (idempotent)
	if err := db.Close(); err != nil {
		// It's okay if it returns an error on second close
		_ = err
	}
}

func TestDatabase_PoolConfiguration(t *testing.T) {
	logger := slog.Default()

	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		DBName:          "anki",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    3,
		ConnMaxLifetime: 10,
	}

	db, err := NewDatabase(cfg, logger)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
	}
	defer db.Close()

	// Verify pool settings
	stats := db.DB.Stats()
	
	// MaxOpenConnections should match config
	if stats.MaxOpenConnections != cfg.MaxConnections {
		t.Errorf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, cfg.MaxConnections)
	}

	// Note: MaxIdleConnections is not directly accessible via Stats,
	// but we can verify the connection works
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping() after pool config error = %v", err)
	}
}
