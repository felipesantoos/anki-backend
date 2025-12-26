package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// PostgresRepository wraps the sql.DB connection with additional functionality
// Implements IDatabaseRepository interface
type PostgresRepository struct {
	DB *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL connection with connection pooling configured
func NewPostgresRepository(cfg config.DatabaseConfig, logger *slog.Logger) (*PostgresRepository, error) {
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
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

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
		"connection_max_idle_time_minutes", cfg.ConnMaxIdleTime,
	)

	return &PostgresRepository{DB: db}, nil
}

// Ping verifies the database connection
// Implements IDatabaseRepository interface
func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

// Close closes the database connection gracefully
func (p *PostgresRepository) Close() error {
	return p.DB.Close()
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

// Ensure PostgresRepository implements IDatabaseRepository
var _ secondary.IDatabaseRepository = (*PostgresRepository)(nil)

