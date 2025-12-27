package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/routes"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	postgresInfra "github.com/felipesantos/anki-backend/infra/postgres"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func setupAuthTestDB(t *testing.T) (*postgresInfra.PostgresRepository, func()) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	log := logger.GetLogger()
	
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
	
	// Get the absolute path to migrations directory relative to this test file
	// tests/integration/auth_test.go -> ../../migrations
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	migrationsDir := filepath.Join(testDir, "..", "..", "migrations")
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
	repoDB, err := postgresInfra.NewPostgresRepository(cfg.Database, log)
	if err != nil {
		t.Skipf("Database not available: %v", err)
	}

	cleanup := func() {
		// Clean up test data
		_, err := repoDB.DB.Exec("DELETE FROM decks WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'test%@example.com')")
		if err != nil {
			t.Logf("Failed to clean up decks: %v", err)
		}
		
		_, err = repoDB.DB.Exec("DELETE FROM users WHERE email LIKE 'test%@example.com'")
		if err != nil {
			t.Logf("Failed to clean up users: %v", err)
		}
		
		repoDB.Close()
	}

	return repoDB, cleanup
}

func TestAuth_Register_Integration(t *testing.T) {
	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	// Setup event bus
	log := logger.GetLogger()
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err := eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	// Setup repositories and service
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	authSvc := authService.NewAuthService(userRepo, deckRepo, eventBus)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := request.RegisterRequest{
			Email:           "testuser@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var result response.RegisterResponse
		err = json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.NotZero(t, result.User.ID)
		assert.Equal(t, "testuser@example.com", result.User.Email)
		assert.False(t, result.User.EmailVerified)
		assert.NotZero(t, result.User.CreatedAt)

		// Verify deck was created
		var deckCount int
		err = db.DB.QueryRow("SELECT COUNT(*) FROM decks WHERE user_id = $1 AND name = 'Default'", result.User.ID).Scan(&deckCount)
		require.NoError(t, err)
		assert.Equal(t, 1, deckCount)
	})

	t.Run("duplicate email", func(t *testing.T) {
		// Register first user
		reqBody1 := request.RegisterRequest{
			Email:           "testduplicate@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		jsonBody1, _ := json.Marshal(reqBody1)
		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody1))
		req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec1 := httptest.NewRecorder()
		e.ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusCreated, rec1.Code)

		// Try to register again with same email
		jsonBody2, _ := json.Marshal(reqBody1)
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody2))
		req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req2)

		assert.Equal(t, http.StatusConflict, rec2.Code)
	})

	t.Run("invalid email format", func(t *testing.T) {
		reqBody := request.RegisterRequest{
			Email:           "invalid-email",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("password too short", func(t *testing.T) {
		reqBody := request.RegisterRequest{
			Email:           "testshortpass@example.com",
			Password:        "pass1",
			PasswordConfirm: "pass1",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("password mismatch", func(t *testing.T) {
		reqBody := request.RegisterRequest{
			Email:           "testmismatch@example.com",
			Password:        "password123",
			PasswordConfirm: "password456",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
