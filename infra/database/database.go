package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/felipesantos/anki-backend/config"
)

// Database wraps the sql.DB connection with additional functionality
type Database struct {
	DB *sql.DB
}

// NewDatabase creates a new database connection with connection pooling configured
func NewDatabase(cfg config.DatabaseConfig, logger *slog.Logger) (*Database, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	logger.Info("Connecting to database",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
		"user", cfg.User,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute) // Default 10 minutes

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		"max_connections", cfg.MaxConnections,
		"max_idle_connections", cfg.MaxIdleConns,
		"connection_max_lifetime_minutes", cfg.ConnMaxLifetime,
	)

	return &Database{DB: db}, nil
}

// Ping verifies the database connection
func (d *Database) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}

// Close closes the database connection gracefully
func (d *Database) Close() error {
	return d.DB.Close()
}

// buildDSN builds a PostgreSQL connection string (DSN) from DatabaseConfig
func buildDSN(cfg config.DatabaseConfig) (string, error) {
	// URL encode password to handle special characters
	password := url.QueryEscape(cfg.Password)

	// Build DSN in format: postgres://user:password@host:port/dbname?sslmode=mode
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	return dsn, nil
}
