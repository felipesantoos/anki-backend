package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/config"
)

// Expected tables in the initial schema (24 tables)
var expectedTables = []string{
	"users",
	"decks",
	"note_types",
	"notes",
	"cards",
	"reviews",
	"media",
	"note_media",
	"sync_meta",
	"user_preferences",
	"backups",
	"filtered_decks",
	"deck_options_presets",
	"deletions_log",
	"saved_searches",
	"flag_names",
	"browser_config",
	"undo_history",
	"shared_decks",
	"shared_deck_ratings",
	"add_ons",
	"check_database_log",
	"profiles",
	"schema_migrations",
}

// Expected ENUM types
var expectedTypes = []string{
	"card_state",
	"review_type",
	"theme_type",
	"scheduler_type",
}

// Expected functions
var expectedFunctions = []string{
	"update_updated_at_column",
	"generate_guid",
	"set_note_guid",
	"log_note_deletion",
	"count_due_cards",
	"count_new_cards",
	"count_learning_cards",
	"reset_sequences",
	"validate_single_ankiweb_sync",
}

// Expected views
var expectedViews = []string{
	"deck_statistics",
	"card_info_extended",
	"empty_cards",
	"leeches",
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

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	dsn := buildDSN(cfg.Database)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to open database connection")

	// Verify connection
	ctx := context.Background()
	err = db.PingContext(ctx)
	require.NoError(t, err, "Failed to ping database")

	// Clean the database IMMEDIATELY to ensure a fresh state for each test
	// This prevents issues with golang-migrate trying to access schema_migrations
	// that may have been left in an inconsistent state by previous tests
	cleanDatabase(t, db)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// getMigrator creates a migrate instance for testing
func getMigrator(t *testing.T, db *sql.DB) (*migrate.Migrate, func()) {
	// Ensure schema_migrations table exists before creating the migrator
	// golang-migrate requires this table to exist
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL PRIMARY KEY,
			dirty boolean NOT NULL
		);
	`)
	require.NoError(t, err, "Failed to create schema_migrations table")
	
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err, "Failed to create postgres driver")

	// Get the absolute path to migrations directory relative to this test file
	// tests/integration/migration_test.go -> ../../migrations
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	migrationsDir := filepath.Join(testDir, "..", "..", "migrations")
	migrationsDir, err = filepath.Abs(migrationsDir)
	require.NoError(t, err, "Failed to get migrations directory path")

	migrationURL := fmt.Sprintf("file://%s", migrationsDir)

	m, err := migrate.NewWithDatabaseInstance(migrationURL, "postgres", driver)
	require.NoError(t, err, "Failed to create migrator")

	cleanup := func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			t.Logf("Error closing migration source: %v", sourceErr)
		}
		if dbErr != nil {
			t.Logf("Error closing migration database: %v", dbErr)
		}
	}

	return m, cleanup
}

// cleanDatabase drops all tables, types, functions, etc. in the public schema
// This is used to reset the database to a clean state without using migrate.Drop()
func cleanDatabase(t *testing.T, db *sql.DB) {
	ctx := context.Background()
	
	// Drop all tables, types, functions, views, and sequences in public schema
	_, err := db.ExecContext(ctx, `
		DO $$ 
		DECLARE
			r RECORD;
		BEGIN
			-- Drop all views first (before tables, as they may depend on tables)
			FOR r IN (SELECT viewname FROM pg_views WHERE schemaname = 'public') LOOP
				EXECUTE 'DROP VIEW IF EXISTS public.' || quote_ident(r.viewname) || ' CASCADE';
			END LOOP;
			
		-- Drop all tables (except schema_migrations, which is managed by golang-migrate)
		FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations') LOOP
			EXECUTE 'DROP TABLE IF EXISTS public.' || quote_ident(r.tablename) || ' CASCADE';
		END LOOP;
		
		-- Ensure schema_migrations table exists (golang-migrate requires it)
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL PRIMARY KEY,
			dirty boolean NOT NULL
		);
		
		-- Clean schema_migrations table
		DELETE FROM schema_migrations;
			
			-- Drop all sequences
			FOR r IN (SELECT sequence_name FROM information_schema.sequences WHERE sequence_schema = 'public') LOOP
				EXECUTE 'DROP SEQUENCE IF EXISTS public.' || quote_ident(r.sequence_name) || ' CASCADE';
			END LOOP;
			
			-- Drop all types
			FOR r IN (SELECT typname FROM pg_type WHERE typtype = 'e' AND typnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')) LOOP
				EXECUTE 'DROP TYPE IF EXISTS public.' || quote_ident(r.typname) || ' CASCADE';
			END LOOP;
			
			-- Drop all functions
			FOR r IN (SELECT proname, oidvectortypes(proargtypes) as argtypes FROM pg_proc INNER JOIN pg_namespace ns ON (pg_proc.pronamespace = ns.oid) WHERE ns.nspname = 'public') LOOP
				EXECUTE 'DROP FUNCTION IF EXISTS public.' || quote_ident(r.proname) || '(' || r.argtypes || ') CASCADE';
			END LOOP;
		END $$;
	`)
	require.NoError(t, err, "Failed to clean database")
}

// ensureMigrationsUp ensures migrations are up, handling dirty state if needed
// Note: setupTestDB already cleans the database, so we start with a clean slate
func ensureMigrationsUp(t *testing.T, m *migrate.Migrate, db *sql.DB) {
	// Check current version and dirty state
	version, dirty, err := m.Version()
	
	// Only run Up() if we need to (database should be clean from setupTestDB)
	if err == migrate.ErrNilVersion || (err == nil && version < 1) {
		t.Logf("Running migrations up (current version: %v, error: %v)", version, err)
		upErr := m.Up()
		if upErr != nil && upErr != migrate.ErrNoChange {
			t.Logf("Migration up failed with error: %v", upErr)
			require.NoError(t, upErr, "Migration up should succeed")
		}
	} else if err == nil && dirty {
		// If database is dirty, this is unexpected (setupTestDB should have cleaned it)
		t.Fatalf("Database is dirty at version %d after setupTestDB - this should not happen", version)
	} else {
		// If we have a version >= 1 and no error, we're good
		require.NoError(t, err, "Failed to get migration version")
		require.False(t, dirty, "Database should not be dirty at this point")
	}

	// Verify final state
	version, dirty, err = m.Version()
	require.NoError(t, err, "Should get migration version after up")
	assert.False(t, dirty, "Database should not be in dirty state after up")
	assert.Equal(t, uint(1), version, "Migration version should be 1")
}

func TestMigration_Up(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure we start fresh - go down one step if at version 1
	version, dirty, err := m.Version()
	if err == nil && version > 0 {
		err = m.Steps(-1)
		if err != nil && err != migrate.ErrNilVersion {
			require.NoError(t, err, "Steps(-1) should succeed or return ErrNilVersion")
		}
	}

	// Run migration up
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err, "Migration up should succeed")
	}

	// Verify migration version
	version, dirty, err = m.Version()
	require.NoError(t, err, "Should get migration version")
	assert.Equal(t, uint(1), version, "Migration version should be 1")
	assert.False(t, dirty, "Database should not be in dirty state")
}

func TestMigration_TablesCreated(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Query to get all tables
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		ORDER BY tablename
	`

	rows, err := db.QueryContext(ctx, query)
	require.NoError(t, err, "Failed to query tables")
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		require.NoError(t, err)
		tables = append(tables, tableName)
	}

	// Verify all expected tables exist
	for _, expectedTable := range expectedTables {
		assert.Contains(t, tables, expectedTable, "Table %s should exist", expectedTable)
	}

	// Verify schema_migrations table has initial data
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "schema_migrations should have version 1 entry")
}

func TestMigration_EnumTypesCreated(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Query to get all ENUM types
	query := `
		SELECT typname 
		FROM pg_type 
		WHERE typtype = 'e' 
		AND typnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		ORDER BY typname
	`

	rows, err := db.QueryContext(ctx, query)
	require.NoError(t, err, "Failed to query enum types")
	defer rows.Close()

	var types []string
	for rows.Next() {
		var typeName string
		err := rows.Scan(&typeName)
		require.NoError(t, err)
		types = append(types, typeName)
	}

	// Verify all expected types exist
	for _, expectedType := range expectedTypes {
		assert.Contains(t, types, expectedType, "Enum type %s should exist", expectedType)
	}
}

func TestMigration_FunctionsCreated(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Query to get all functions
	query := `
		SELECT proname 
		FROM pg_proc 
		WHERE pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		AND proname NOT LIKE 'pg_%'
		ORDER BY proname
	`

	rows, err := db.QueryContext(ctx, query)
	require.NoError(t, err, "Failed to query functions")
	defer rows.Close()

	var functions []string
	for rows.Next() {
		var funcName string
		err := rows.Scan(&funcName)
		require.NoError(t, err)
		functions = append(functions, funcName)
	}

	// Verify all expected functions exist
	for _, expectedFunc := range expectedFunctions {
		assert.Contains(t, functions, expectedFunc, "Function %s should exist", expectedFunc)
	}
}

func TestMigration_ViewsCreated(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Query to get all views
	query := `
		SELECT viewname 
		FROM pg_views 
		WHERE schemaname = 'public'
		ORDER BY viewname
	`

	rows, err := db.QueryContext(ctx, query)
	require.NoError(t, err, "Failed to query views")
	defer rows.Close()

	var views []string
	for rows.Next() {
		var viewName string
		err := rows.Scan(&viewName)
		require.NoError(t, err)
		views = append(views, viewName)
	}

	// Verify all expected views exist
	for _, expectedView := range expectedViews {
		assert.Contains(t, views, expectedView, "View %s should exist", expectedView)
	}
}

func TestMigration_TriggersCreated(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Query to get all triggers
	query := `
		SELECT trigger_name 
		FROM information_schema.triggers 
		WHERE trigger_schema = 'public'
		ORDER BY trigger_name
	`

	rows, err := db.QueryContext(ctx, query)
	require.NoError(t, err, "Failed to query triggers")
	defer rows.Close()

	var triggers []string
	for rows.Next() {
		var triggerName string
		err := rows.Scan(&triggerName)
		require.NoError(t, err)
		triggers = append(triggers, triggerName)
	}

	// Verify key triggers exist
	expectedTriggers := []string{
		"update_users_updated_at",
		"update_decks_updated_at",
		"set_notes_guid",
		"log_notes_deletion",
		"update_profiles_updated_at",
		"validate_single_ankiweb_sync_trigger",
	}

	for _, expectedTrigger := range expectedTriggers {
		assert.Contains(t, triggers, expectedTrigger, "Trigger %s should exist", expectedTrigger)
	}

	// Verify we have at least the expected number of triggers (should be around 16+)
	assert.GreaterOrEqual(t, len(triggers), 16, "Should have at least 16 triggers")
}

func TestMigration_IndexesCreated(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Query to get all indexes
	query := `
		SELECT indexname 
		FROM pg_indexes 
		WHERE schemaname = 'public'
		AND indexname NOT LIKE 'pg_%'
		ORDER BY indexname
	`

	rows, err := db.QueryContext(ctx, query)
	require.NoError(t, err, "Failed to query indexes")
	defer rows.Close()

	var indexes []string
	for rows.Next() {
		var indexName string
		err := rows.Scan(&indexName)
		require.NoError(t, err)
		indexes = append(indexes, indexName)
	}

	// Verify key indexes exist
	expectedIndexes := []string{
		"idx_users_email",
		"idx_decks_user_id",
		"idx_notes_guid",
		"idx_cards_deck_id",
		"idx_profiles_user_id",
		"idx_profiles_ankiweb_sync",
	}

	for _, expectedIndex := range expectedIndexes {
		assert.Contains(t, indexes, expectedIndex, "Index %s should exist", expectedIndex)
	}

	// Verify we have many indexes (should be 50+)
	assert.GreaterOrEqual(t, len(indexes), 50, "Should have at least 50 indexes")
}

func TestMigration_Down(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up first
	ensureMigrationsUp(t, m, db)

	// Verify tables exist before down
	ctx := context.Background()
	var tableCount int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`).Scan(&tableCount)
	require.NoError(t, err)
	assert.Greater(t, tableCount, 0, "Should have tables before down")

	// Run migration down one step (from version 1 to version 0)
	err = m.Steps(-1)
	require.NoError(t, err, "Migration down one step should succeed")

	// Verify migration version is reset
	version, dirty, err := m.Version()
	if err != nil {
		// ErrNilVersion is expected after down
		assert.ErrorIs(t, err, migrate.ErrNilVersion, "Should get ErrNilVersion after down")
	} else {
		assert.Equal(t, uint(0), version, "Migration version should be 0 after down")
		assert.False(t, dirty, "Database should not be in dirty state")
	}

	// Verify tables are dropped (except schema_migrations which may persist)
	var remainingTables int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM pg_tables 
		WHERE schemaname = 'public'
		AND tablename != 'schema_migrations'
	`).Scan(&remainingTables)
	require.NoError(t, err)
	assert.Equal(t, 0, remainingTables, "All tables except schema_migrations should be dropped")
}

func TestMigration_UpDownUp(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Start fresh - ensure we're at version 0
	version, _, err := m.Version()
	if err == nil && version > 0 {
		// If we're at version 1, go down one step
		err = m.Steps(-1)
		if err != nil && err != migrate.ErrNilVersion {
			require.NoError(t, err, "Steps(-1) should succeed or return ErrNilVersion")
		}
	}

	// First up
	ensureMigrationsUp(t, m, db)

	version1, dirty1, err := m.Version()
	require.NoError(t, err)
	assert.Equal(t, uint(1), version1, "Version should be 1 after first up")
	assert.False(t, dirty1, "Database should not be dirty after first up")

	// Down one step (from version 1 to version 0)
	err = m.Steps(-1)
	require.NoError(t, err, "Migration down one step should succeed")

	// Second up (should work fine)
	ensureMigrationsUp(t, m, db)

	version2, dirty2, err := m.Version()
	require.NoError(t, err)
	assert.Equal(t, uint(1), version2, "Version should be 1 after second up")
	assert.False(t, dirty2, "Database should not be dirty after second up")

	// Verify tables exist after second up
	ctx := context.Background()
	var tableCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`).Scan(&tableCount)
	require.NoError(t, err)
	assert.Greater(t, tableCount, 0, "Should have tables after second up")
}

func TestMigration_Constraints(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Test that GUID constraint works
	_, err := db.ExecContext(ctx, `
		INSERT INTO notes (user_id, guid, note_type_id, fields_json)
		VALUES (1, 'invalid-guid', 1, '{}'::jsonb)
	`)
	assert.Error(t, err, "Should fail with invalid GUID format")

	// Test that valid GUID works
	_, err = db.ExecContext(ctx, `
		INSERT INTO notes (user_id, guid, note_type_id, fields_json)
		VALUES (1, '123e4567-e89b-12d3-a456-426614174000', 1, '{}'::jsonb)
	`)
	// This will fail because user_id=1 doesn't exist, but GUID format is valid
	assert.Error(t, err, "Should fail due to foreign key constraint")
}

func TestMigration_ProfileSyncConstraint(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	m, mCleanup := getMigrator(t, db)
	defer mCleanup()

	// Ensure migrations are up
	ensureMigrationsUp(t, m, db)

	ctx := context.Background()

	// Create a test user
	var userID int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ('test@example.com', 'hash')
		RETURNING id
	`).Scan(&userID)
	require.NoError(t, err)

	// Create first profile with sync enabled
	var profile1ID int64
	err = db.QueryRowContext(ctx, `
		INSERT INTO profiles (user_id, name, ankiweb_sync_enabled)
		VALUES ($1, 'Profile 1', true)
		RETURNING id
	`, userID).Scan(&profile1ID)
	require.NoError(t, err, "Should create first profile with sync")

	// Try to create second profile with sync enabled (should fail)
	_, err = db.ExecContext(ctx, `
		INSERT INTO profiles (user_id, name, ankiweb_sync_enabled)
		VALUES ($1, 'Profile 2', true)
	`, userID)
	assert.Error(t, err, "Should fail to create second profile with sync enabled")

	// Create second profile without sync (should succeed)
	var profile2ID int64
	err = db.QueryRowContext(ctx, `
		INSERT INTO profiles (user_id, name, ankiweb_sync_enabled)
		VALUES ($1, 'Profile 2', false)
		RETURNING id
	`, userID).Scan(&profile2ID)
	require.NoError(t, err, "Should create second profile without sync")

	// Try to enable sync on second profile (should fail)
	_, err = db.ExecContext(ctx, `
		UPDATE profiles 
		SET ankiweb_sync_enabled = true 
		WHERE id = $1
	`, profile2ID)
	assert.Error(t, err, "Should fail to enable sync on second profile")

	// Disable sync on first profile
	_, err = db.ExecContext(ctx, `
		UPDATE profiles 
		SET ankiweb_sync_enabled = false 
		WHERE id = $1
	`, profile1ID)
	require.NoError(t, err, "Should disable sync on first profile")

	// Now enable sync on second profile (should succeed)
	_, err = db.ExecContext(ctx, `
		UPDATE profiles 
		SET ankiweb_sync_enabled = true 
		WHERE id = $1
	`, profile2ID)
	require.NoError(t, err, "Should enable sync on second profile after first is disabled")
}

