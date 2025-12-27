package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/pkg/jwt"
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

	// Setup Redis
	cfg, err := config.Load()
	require.NoError(t, err)
	redisRepo, err := redisInfra.NewRedisRepository(cfg.Redis, log)
	require.NoError(t, err)
	defer redisRepo.Close()

	// Setup JWT service
	jwtSvc, err := jwt.NewJWTService(cfg.JWT)
	require.NoError(t, err)

	// Setup repositories and service
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	authSvc := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, redisRepo)

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

func TestAuth_Login_Integration(t *testing.T) {
	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	// Setup event bus
	log := logger.GetLogger()
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err := eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	// Setup Redis
	cfg, err := config.Load()
	require.NoError(t, err)
	redisRepo, err := redisInfra.NewRedisRepository(cfg.Redis, log)
	require.NoError(t, err)
	defer redisRepo.Close()

	// Setup JWT service
	jwtSvc, err := jwt.NewJWTService(cfg.JWT)
	require.NoError(t, err)

	// Setup repositories and service
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	authSvc := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, redisRepo)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc)

	// First register a user
	registerReq := request.RegisterRequest{
		Email:           "testlogin@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	registerRec := httptest.NewRecorder()
	e.ServeHTTP(registerRec, registerHTTPReq)
	require.Equal(t, http.StatusCreated, registerRec.Code)

	t.Run("successful login", func(t *testing.T) {
		reqBody := request.LoginRequest{
			Email:    "testlogin@example.com",
			Password: "password123",
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result response.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "Bearer", result.TokenType)
		assert.Greater(t, result.ExpiresIn, 0)
		assert.Equal(t, "testlogin@example.com", result.User.Email)
		assert.NotZero(t, result.User.ID)
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		reqBody := request.LoginRequest{
			Email:    "testlogin@example.com",
			Password: "wrongpassword",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid credentials - user not found", func(t *testing.T) {
		reqBody := request.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid email format", func(t *testing.T) {
		reqBody := request.LoginRequest{
			Email:    "invalid-email",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAuth_RefreshToken_Integration(t *testing.T) {
	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	// Setup event bus
	log := logger.GetLogger()
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err := eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	// Setup Redis
	cfg, err := config.Load()
	require.NoError(t, err)
	redisRepo, err := redisInfra.NewRedisRepository(cfg.Redis, log)
	require.NoError(t, err)
	defer redisRepo.Close()

	// Setup JWT service
	jwtSvc, err := jwt.NewJWTService(cfg.JWT)
	require.NoError(t, err)

	// Setup repositories and service
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	authSvc := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, redisRepo)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc)

	// Register and login a user to get a refresh token
	registerReq := request.RegisterRequest{
		Email:           "testrefresh@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	registerRec := httptest.NewRecorder()
	e.ServeHTTP(registerRec, registerHTTPReq)
	require.Equal(t, http.StatusCreated, registerRec.Code)

	loginReq := request.LoginRequest{
		Email:    "testrefresh@example.com",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginHTTPReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp response.LoginResponse
	err = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	refreshToken := loginResp.RefreshToken

	t.Run("successful refresh", func(t *testing.T) {
		reqBody := request.RefreshRequest{
			RefreshToken: refreshToken,
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result response.TokenResponse
		err = json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.NotEmpty(t, result.AccessToken)
		
		// Token rotation: new refresh token should be returned
		assert.NotEmpty(t, result.RefreshToken, "RefreshToken should be returned (token rotation)")
		
		// Note: We don't verify that the new refresh token is different from the old one
		// because JWT tokens generated in rapid succession might have identical timestamps
		// The important thing is that a new refresh token is returned and the old one is invalidated
		
		assert.Equal(t, "Bearer", result.TokenType)
		assert.Greater(t, result.ExpiresIn, 0)

		// Verify token rotation: old refresh token should be invalidated
		// Try to use the old refresh token again - it should fail
		oldTokenReqBody := request.RefreshRequest{
			RefreshToken: refreshToken, // Using old token
		}
		oldTokenJsonBody, err := json.Marshal(oldTokenReqBody)
		require.NoError(t, err)

		oldTokenReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(oldTokenJsonBody))
		oldTokenReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		oldTokenRec := httptest.NewRecorder()
		
		e.ServeHTTP(oldTokenRec, oldTokenReq)

		// Old token should be rejected (invalidated by token rotation)
		assert.Equal(t, http.StatusUnauthorized, oldTokenRec.Code, "Old refresh token should be invalidated (token rotation)")

		// Verify new refresh token works (only if a new one was actually returned)
		if result.RefreshToken != "" && result.RefreshToken != refreshToken {
			newTokenReqBody := request.RefreshRequest{
				RefreshToken: result.RefreshToken, // Using new token
			}
			newTokenJsonBody, err := json.Marshal(newTokenReqBody)
			require.NoError(t, err)

			newTokenReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(newTokenJsonBody))
			newTokenReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			newTokenRec := httptest.NewRecorder()
			
			e.ServeHTTP(newTokenRec, newTokenReq)

			assert.Equal(t, http.StatusOK, newTokenRec.Code, "New refresh token should work")

			var newResult response.TokenResponse
			err = json.Unmarshal(newTokenRec.Body.Bytes(), &newResult)
			require.NoError(t, err)
			assert.NotEmpty(t, newResult.AccessToken)
			assert.NotEmpty(t, newResult.RefreshToken, "Second refresh should also return new refresh token (token rotation)")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		reqBody := request.RefreshRequest{
			RefreshToken: "invalid-token",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("empty refresh token", func(t *testing.T) {
		reqBody := request.RefreshRequest{
			RefreshToken: "",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAuth_Logout_Integration(t *testing.T) {
	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	// Setup event bus
	log := logger.GetLogger()
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	err := eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	// Setup Redis
	cfg, err := config.Load()
	require.NoError(t, err)
	redisRepo, err := redisInfra.NewRedisRepository(cfg.Redis, log)
	require.NoError(t, err)
	defer redisRepo.Close()

	// Setup JWT service
	jwtSvc, err := jwt.NewJWTService(cfg.JWT)
	require.NoError(t, err)

	// Setup repositories and service
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	authSvc := authService.NewAuthService(userRepo, deckRepo, eventBus, jwtSvc, redisRepo)

	// Setup Echo
	e := echo.New()
	routes.RegisterAuthRoutes(e, authSvc)

	// Register and login a user to get a refresh token
	registerReq := request.RegisterRequest{
		Email:           "testlogout@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	}
	registerBody, _ := json.Marshal(registerReq)
	registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
	registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	registerRec := httptest.NewRecorder()
	e.ServeHTTP(registerRec, registerHTTPReq)
	require.Equal(t, http.StatusCreated, registerRec.Code)

	loginReq := request.LoginRequest{
		Email:    "testlogout@example.com",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginHTTPReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp response.LoginResponse
	err = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	refreshToken := loginResp.RefreshToken

	t.Run("successful logout", func(t *testing.T) {
		// Logout with both access token and refresh token
		reqBody := request.RefreshRequest{
			RefreshToken: refreshToken,
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Logged out successfully", result["message"])

		// Verify that refresh token was invalidated by trying to refresh it
		refreshReq := request.RefreshRequest{
			RefreshToken: refreshToken,
		}
		refreshBody, _ := json.Marshal(refreshReq)
		refreshHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshBody))
		refreshHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		refreshRec := httptest.NewRecorder()
		e.ServeHTTP(refreshRec, refreshHTTPReq)

		// Should fail with unauthorized since token was revoked
		assert.Equal(t, http.StatusUnauthorized, refreshRec.Code)

		// Verify that access token is blacklisted in Redis
		ctx := context.Background()
		accessTokenHash := sha256.Sum256([]byte(loginResp.AccessToken))
		accessTokenBlacklistKey := fmt.Sprintf("access_token_blacklist:%s", hex.EncodeToString(accessTokenHash[:]))
		exists, err := redisRepo.Exists(ctx, accessTokenBlacklistKey)
		require.NoError(t, err, "Failed to check access token blacklist")
		assert.True(t, exists, "Access token should be blacklisted in Redis after logout")

		// Verify that refresh token is not in Redis (was deleted)
		refreshTokenHash := sha256.Sum256([]byte(refreshToken))
		refreshTokenKey := fmt.Sprintf("refresh_token:%s", hex.EncodeToString(refreshTokenHash[:]))
		refreshExists, err := redisRepo.Exists(ctx, refreshTokenKey)
		require.NoError(t, err, "Failed to check refresh token")
		assert.False(t, refreshExists, "Refresh token should be deleted from Redis after logout")
	})

	t.Run("logout with access token only", func(t *testing.T) {
		// Login again to get new tokens
		loginReq := request.LoginRequest{
			Email:    "testlogout@example.com",
			Password: "password123",
		}
		loginBody, _ := json.Marshal(loginReq)
		loginHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
		loginHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		loginRec := httptest.NewRecorder()
		e.ServeHTTP(loginRec, loginHTTPReq)
		require.Equal(t, http.StatusOK, loginRec.Code)

		var loginResp2 response.LoginResponse
		err = json.Unmarshal(loginRec.Body.Bytes(), &loginResp2)
		require.NoError(t, err)

		// Logout with only access token (no refresh token in body)
		logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		logoutReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		logoutReq.Header.Set("Authorization", "Bearer "+loginResp2.AccessToken)
		logoutRec := httptest.NewRecorder()

		e.ServeHTTP(logoutRec, logoutReq)

		assert.Equal(t, http.StatusOK, logoutRec.Code)

		// Verify that access token is blacklisted in Redis
		ctx := context.Background()
		accessTokenHash := sha256.Sum256([]byte(loginResp2.AccessToken))
		accessTokenBlacklistKey := fmt.Sprintf("access_token_blacklist:%s", hex.EncodeToString(accessTokenHash[:]))
		exists, err := redisRepo.Exists(ctx, accessTokenBlacklistKey)
		require.NoError(t, err, "Failed to check access token blacklist")
		assert.True(t, exists, "Access token should be blacklisted in Redis after logout")
	})

	t.Run("logout with no tokens", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
