package database

import (
	"log/slog"
	"net/url"
	"testing"

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

// Note: Tests that require real database connections (Ping, Close, PoolConfiguration)
// are located in tests/integration/database_test.go to avoid duplication and
// keep unit tests in infra/ focused on unit-level testing without external dependencies.
