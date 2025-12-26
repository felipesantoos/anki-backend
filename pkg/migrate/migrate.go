package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/felipesantos/anki-backend/config"
)

const migrationsPath = "migrations"

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

// RunMigrations executes all pending database migrations
func RunMigrations(cfg config.DatabaseConfig, logger *slog.Logger) error {
	// Build DSN
	dsn, err := buildDSN(cfg)
	if err != nil {
		return fmt.Errorf("failed to build DSN: %w", err)
	}

	logger.Info("Starting database migrations",
		"database", cfg.DBName,
	)

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path to migrations directory
	migrationsDir, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migrations directory path: %w", err)
	}

	// Create migrator with file:// source
	migrationURL := fmt.Sprintf("file://%s", migrationsDir)
	m, err := migrate.NewWithDatabaseInstance(
		migrationURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Get current version
	version, dirty, err := m.Version()
	currentVersion := uint(0)
	hasVersion := false

	if err != nil {
		if !errors.Is(err, migrate.ErrNilVersion) {
			return fmt.Errorf("failed to get current version: %w", err)
		}
		logger.Info("No migrations applied yet, database is empty")
	} else {
		hasVersion = true
		currentVersion = version
		if dirty {
			logger.Warn("Database is in dirty state, migrations may have failed previously",
				"version", version,
			)
			return fmt.Errorf("database is in dirty state at version %d, run 'migrate force' to fix", version)
		}
		logger.Info("Current migration version",
			"version", version,
		)
	}

	// Run migrations
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("No pending migrations, database is up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get new version
	newVersion, _, err := m.Version()
	if err != nil {
		// If we get ErrNilVersion, it means no migrations were applied
		// This shouldn't happen after m.Up(), but handle it gracefully
		if errors.Is(err, migrate.ErrNilVersion) {
			logger.Info("Migrations completed, but no version found (unexpected)")
			return nil
		}
		return fmt.Errorf("failed to get new version: %w", err)
	}

	var prevVersion uint = 0
	if hasVersion {
		prevVersion = currentVersion
	}

	logger.Info("Migrations completed successfully",
		"previous_version", prevVersion,
		"new_version", newVersion,
	)

	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(cfg config.DatabaseConfig, logger *slog.Logger) (uint, bool, error) {
	// Build DSN
	dsn, err := buildDSN(cfg)
	if err != nil {
		return 0, false, fmt.Errorf("failed to build DSN: %w", err)
	}

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path to migrations directory
	migrationsDir, err := filepath.Abs(migrationsPath)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migrations directory path: %w", err)
	}

	// Create migrator with file:// source
	migrationURL := fmt.Sprintf("file://%s", migrationsDir)
	m, err := migrate.NewWithDatabaseInstance(
		migrationURL,
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil // No migrations applied (version=0, dirty=false, no error)
		}
		return 0, false, err
	}

	return version, dirty, nil
}
