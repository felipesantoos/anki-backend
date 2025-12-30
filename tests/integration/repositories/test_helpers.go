package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	postgresInfra "github.com/felipesantos/anki-backend/infra/postgres"
)

// setupTestDB creates a test database connection with migrations
func setupTestDB(t *testing.T) (*postgresInfra.PostgresRepository, func()) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	// Create database connection for migrations
	dsn := buildDSN(cfg.Database)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to open database connection")

	// Ensure schema_migrations table exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL PRIMARY KEY,
			dirty boolean NOT NULL
		);
	`)
	require.NoError(t, err, "Failed to ensure schema_migrations table exists")

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err, "Failed to create postgres driver")

	// Get the absolute path to migrations directory
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	migrationsDir := filepath.Join(testDir, "..", "..", "..", "migrations")
	migrationsDir, err = filepath.Abs(migrationsDir)
	require.NoError(t, err, "Failed to get migrations directory path")

	migrationURL := fmt.Sprintf("file://%s", migrationsDir)
	m, err := migrate.NewWithDatabaseInstance(migrationURL, "postgres", driver)
	require.NoError(t, err, "Failed to create migrator")

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err, "Failed to run migrations")
	}

	// Close migrator
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		t.Logf("Error closing migration source: %v", sourceErr)
	}
	if dbErr != nil {
		t.Logf("Error closing migration database: %v", dbErr)
	}

	// Close the temporary DB connection
	db.Close()

	// Create PostgresRepository for the test
	repoDB, err := postgresInfra.NewPostgresRepository(cfg.Database, nil)
	if err != nil {
		t.Skipf("Database not available: %v", err)
	}

	cleanup := func() {
		// Clean up test data in reverse order of dependencies
		// Use TRUNCATE CASCADE for a fresh start or delete carefully
		tables := []string{
			"reviews",
			"cards",
			"notes",
			"media",
			"decks",
			"filtered_decks",
			"deck_options_presets",
			"note_types",
			"profiles",
			"browser_configs",
			"add_ons",
			"backups",
			"check_database_logs",
			"deletion_logs",
			"flag_names",
			"saved_searches",
			"shared_deck_ratings",
			"shared_decks",
			"sync_metas",
			"undo_histories",
			"user_preferences",
			"users",
		}

		for _, table := range tables {
			_, _ = repoDB.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		}
		repoDB.Close()
	}

	return repoDB, cleanup
}

// buildDSN builds a PostgreSQL connection string (DSN) from DatabaseConfig
func buildDSN(cfg config.DatabaseConfig) string {
	password := url.QueryEscape(cfg.Password)
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// createTestUser creates a test user and returns the user ID and email
func createTestUser(t *testing.T, ctx context.Context, userRepo secondary.IUserRepository, emailSuffix string) (int64, string) {
	uniqueSuffix := fmt.Sprintf("%s_%d", emailSuffix, time.Now().UnixNano())
	emailStr := fmt.Sprintf("test_%s@example.com", uniqueSuffix)
	email, err := valueobjects.NewEmail(emailStr)
	require.NoError(t, err)

	password, err := valueobjects.NewPassword("password123")
	require.NoError(t, err)

	now := time.Now()
	user, err := user.NewBuilder().
		WithID(0).
		WithEmail(email).
		WithPasswordHash(password).
		WithEmailVerified(false).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		WithLastLoginAt(nil).
		WithDeletedAt(nil).
		Build()
	require.NoError(t, err)

	err = userRepo.Save(ctx, user)
	require.NoError(t, err)

	return user.GetID(), emailStr
}

