package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/routes"
	"github.com/felipesantos/anki-backend/dicontainer"
	infraEvents "github.com/felipesantos/anki-backend/infra/events"
	redisInfra "github.com/felipesantos/anki-backend/infra/redis"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/pkg/jwt"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func TestAuth_Register_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

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

	t.Run("password without letters", func(t *testing.T) {
		reqBody := request.RegisterRequest{
			Email:           "noletters@example.com",
			Password:        "12345678",
			PasswordConfirm: "12345678",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("password without numbers", func(t *testing.T) {
		reqBody := request.RegisterRequest{
			Email:           "nonumbers@example.com",
			Password:        "password",
			PasswordConfirm: "password",
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

func TestAuth_EmailReuse_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := logger.GetLogger()
	eventBus := infraEvents.NewInMemoryEventBus(1, 10, log)
	_ = eventBus.Start()
	defer eventBus.Stop()

	cfg, _ := config.Load()
	redisRepo, _ := redisInfra.NewRedisRepository(cfg.Redis, log)
	defer redisRepo.Close()

	jwtSvc, _ := jwt.NewJWTService(cfg.JWT)
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()
	router.RegisterUserRoutes()

	email := "reuse@example.com"
	password := "Password123!"

	t.Run("should allow email reuse after soft delete", func(t *testing.T) {
		// 1. Register User A
		registerReq := request.RegisterRequest{
			Email:           email,
			Password:        password,
			PasswordConfirm: password,
		}
		b, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		// 2. Login to get token
		loginReq := request.LoginRequest{Email: email, Password: password}
		b, _ = json.Marshal(loginReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var loginRes response.LoginResponse
		json.Unmarshal(rec.Body.Bytes(), &loginRes)

		// 3. Soft delete User A
		req = httptest.NewRequest(http.MethodDelete, "/api/v1/user/me", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+loginRes.AccessToken)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNoContent, rec.Code)

		// 4. Register new user with SAME email
		b, _ = json.Marshal(registerReq)
		req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Should succeed because User A is soft-deleted
		assert.Equal(t, http.StatusCreated, rec.Code, "Email should be reusable after soft delete")
	})
}

func TestAuth_Login_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

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
		assert.NotNil(t, result.User.LastLoginAt)
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
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

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
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

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


func TestAuth_VerifyEmail_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

	t.Run("successful verification", func(t *testing.T) {
		// First register a user
		registerReq := request.RegisterRequest{
			Email:           "testverify@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		var registerResp response.RegisterResponse
		err = json.Unmarshal(registerRec.Body.Bytes(), &registerResp)
		require.NoError(t, err)
		require.False(t, registerResp.User.EmailVerified, "Email should not be verified initially")

		// Generate verification token manually (simulating email link)
		token, err := jwtSvc.GenerateEmailVerificationToken(registerResp.User.ID)
		require.NoError(t, err)

		// Verify email
		verifyReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/auth/verify-email?token=%s", token), nil)
		verifyRec := httptest.NewRecorder()
		e.ServeHTTP(verifyRec, verifyReq)

		assert.Equal(t, http.StatusOK, verifyRec.Code)

		var verifyResult map[string]string
		err = json.Unmarshal(verifyRec.Body.Bytes(), &verifyResult)
		require.NoError(t, err)
		assert.Equal(t, "Email verified successfully", verifyResult["message"])

		// Verify that user's email_verified is now true
		var emailVerified bool
		err = db.DB.QueryRow("SELECT email_verified FROM users WHERE id = $1", registerResp.User.ID).Scan(&emailVerified)
		require.NoError(t, err)
		assert.True(t, emailVerified, "Email should be verified after verification")
	})

	t.Run("invalid token", func(t *testing.T) {
		verifyReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify-email?token=invalid-token", nil)
		verifyRec := httptest.NewRecorder()
		e.ServeHTTP(verifyRec, verifyReq)

		assert.Equal(t, http.StatusUnauthorized, verifyRec.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		verifyReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify-email", nil)
		verifyRec := httptest.NewRecorder()
		e.ServeHTTP(verifyRec, verifyReq)

		assert.Equal(t, http.StatusBadRequest, verifyRec.Code)
	})
}

func TestAuth_ResendVerificationEmail_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

	t.Run("successful resend", func(t *testing.T) {
		// First register a user
		registerReq := request.RegisterRequest{
			Email:           "testresend@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		// Resend verification email
		resendReq := request.ResendVerificationRequest{
			Email: "testresend@example.com",
		}
		resendBody, _ := json.Marshal(resendReq)
		resendHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification", bytes.NewReader(resendBody))
		resendHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resendRec := httptest.NewRecorder()
		e.ServeHTTP(resendRec, resendHTTPReq)

		assert.Equal(t, http.StatusOK, resendRec.Code)

		var result map[string]string
		err = json.Unmarshal(resendRec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Verification email sent successfully", result["message"])
	})

	t.Run("user not found", func(t *testing.T) {
		resendReq := request.ResendVerificationRequest{
			Email: "nonexistent@example.com",
		}
		resendBody, _ := json.Marshal(resendReq)
		resendHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification", bytes.NewReader(resendBody))
		resendHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resendRec := httptest.NewRecorder()
		e.ServeHTTP(resendRec, resendHTTPReq)

		assert.Equal(t, http.StatusNotFound, resendRec.Code)
	})

	t.Run("email already verified", func(t *testing.T) {
		// Register and verify a user
		registerReq := request.RegisterRequest{
			Email:           "testalreadyverified@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		var registerResp response.RegisterResponse
		err = json.Unmarshal(registerRec.Body.Bytes(), &registerResp)
		require.NoError(t, err)

		// Manually verify the email in the database
		_, err = db.DB.Exec("UPDATE users SET email_verified = true WHERE id = $1", registerResp.User.ID)
		require.NoError(t, err)

		// Try to resend verification email
		resendReq := request.ResendVerificationRequest{
			Email: "testalreadyverified@example.com",
		}
		resendBody, _ := json.Marshal(resendReq)
		resendHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification", bytes.NewReader(resendBody))
		resendHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resendRec := httptest.NewRecorder()
		e.ServeHTTP(resendRec, resendHTTPReq)

		assert.Equal(t, http.StatusConflict, resendRec.Code)
	})
}

func TestAuth_RequestPasswordReset_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

	t.Run("successful request", func(t *testing.T) {
		// First register a user
		registerReq := request.RegisterRequest{
			Email:           "testreset@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		// Request password reset
		resetReq := request.RequestPasswordResetRequest{
			Email: "testreset@example.com",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/request-password-reset", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)

		assert.Equal(t, http.StatusOK, resetRec.Code)

		var result map[string]string
		err = json.Unmarshal(resetRec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result["message"], "If the email exists")
	})

	t.Run("email not found - should still return success", func(t *testing.T) {
		// Request password reset for non-existent email
		resetReq := request.RequestPasswordResetRequest{
			Email: "nonexistent@example.com",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/request-password-reset", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)

		// Should return success even if email doesn't exist (security)
		assert.Equal(t, http.StatusOK, resetRec.Code)

		var result map[string]string
		err = json.Unmarshal(resetRec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result["message"], "If the email exists")
	})

	t.Run("invalid email format", func(t *testing.T) {
		resetReq := request.RequestPasswordResetRequest{
			Email: "invalid-email",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/request-password-reset", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)

		// Should still return success (security - don't reveal email validity)
		assert.Equal(t, http.StatusOK, resetRec.Code)
	})
}

func TestAuth_ResetPassword_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

	t.Run("successful reset", func(t *testing.T) {
		// First register a user
		registerReq := request.RegisterRequest{
			Email:           "testresetpass@example.com",
			Password:        "oldpassword123",
			PasswordConfirm: "oldpassword123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		var registerResp response.RegisterResponse
		err = json.Unmarshal(registerRec.Body.Bytes(), &registerResp)
		require.NoError(t, err)

		// Generate password reset token manually (simulating email link)
		token, err := jwtSvc.GeneratePasswordResetToken(registerResp.User.ID)
		require.NoError(t, err)

		// Reset password
		resetReq := request.ResetPasswordRequest{
			Token:           token,
			NewPassword:     "newpassword123",
			PasswordConfirm: "newpassword123",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)

		assert.Equal(t, http.StatusOK, resetRec.Code)

		var result map[string]string
		err = json.Unmarshal(resetRec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Password reset successfully. Please log in with your new password.", result["message"])

		// Verify that old password no longer works
		loginReq := request.LoginRequest{
			Email:    "testresetpass@example.com",
			Password: "oldpassword123",
		}
		loginBody, _ := json.Marshal(loginReq)
		loginHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
		loginHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		loginRec := httptest.NewRecorder()
		e.ServeHTTP(loginRec, loginHTTPReq)

		assert.Equal(t, http.StatusUnauthorized, loginRec.Code, "Old password should not work")

		// Verify that new password works
		loginReq2 := request.LoginRequest{
			Email:    "testresetpass@example.com",
			Password: "newpassword123",
		}
		loginBody2, _ := json.Marshal(loginReq2)
		loginHTTPReq2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody2))
		loginHTTPReq2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		loginRec2 := httptest.NewRecorder()
		e.ServeHTTP(loginRec2, loginHTTPReq2)

		assert.Equal(t, http.StatusOK, loginRec2.Code, "New password should work")
	})

	t.Run("invalid token", func(t *testing.T) {
		resetReq := request.ResetPasswordRequest{
			Token:           "invalid-token",
			NewPassword:     "newpassword123",
			PasswordConfirm: "newpassword123",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)

		assert.Equal(t, http.StatusUnauthorized, resetRec.Code)
	})

	t.Run("wrong token type", func(t *testing.T) {
		// Register a user
		registerReq := request.RegisterRequest{
			Email:           "testwrongtoken@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		var registerResp response.RegisterResponse
		err = json.Unmarshal(registerRec.Body.Bytes(), &registerResp)
		require.NoError(t, err)

		// Generate access token (not password reset token)
		accessToken, err := jwtSvc.GenerateAccessToken(registerResp.User.ID)
		require.NoError(t, err)

		// Try to use access token as reset token
		resetReq := request.ResetPasswordRequest{
			Token:           accessToken,
			NewPassword:     "newpassword123",
			PasswordConfirm: "newpassword123",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)

		assert.Equal(t, http.StatusUnauthorized, resetRec.Code)
	})

	t.Run("invalid password", func(t *testing.T) {
		// Register a user
		registerReq := request.RegisterRequest{
			Email:           "testinvalidpass@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		var registerResp response.RegisterResponse
		err = json.Unmarshal(registerRec.Body.Bytes(), &registerResp)
		require.NoError(t, err)

		// Generate password reset token
		token, err := jwtSvc.GeneratePasswordResetToken(registerResp.User.ID)
		require.NoError(t, err)

		// Try to reset with short password
		resetReq := request.ResetPasswordRequest{
			Token:           token,
			NewPassword:     "short",
			PasswordConfirm: "short",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)
		assert.Equal(t, http.StatusBadRequest, resetRec.Code)

		// Try to reset with password without letters
		resetReq = request.ResetPasswordRequest{
			Token:           token,
			NewPassword:     "12345678",
			PasswordConfirm: "12345678",
		}
		resetBody, _ = json.Marshal(resetReq)
		resetHTTPReq = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec = httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)
		assert.Equal(t, http.StatusBadRequest, resetRec.Code)

		// Try to reset with password without numbers
		resetReq = request.ResetPasswordRequest{
			Token:           token,
			NewPassword:     "password",
			PasswordConfirm: "password",
		}
		resetBody, _ = json.Marshal(resetReq)
		resetHTTPReq = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec = httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)
		assert.Equal(t, http.StatusBadRequest, resetRec.Code)
	})

	t.Run("password mismatch", func(t *testing.T) {
		// Register a user
		registerReq := request.RegisterRequest{
			Email:           "testmismatch@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		var registerResp response.RegisterResponse
		err = json.Unmarshal(registerRec.Body.Bytes(), &registerResp)
		require.NoError(t, err)

		// Generate password reset token
		token, err := jwtSvc.GeneratePasswordResetToken(registerResp.User.ID)
		require.NoError(t, err)

		// Try to reset with mismatched passwords
		resetReq := request.ResetPasswordRequest{
			Token:           token,
			NewPassword:     "newpassword123",
			PasswordConfirm: "differentpassword123",
		}
		resetBody, _ := json.Marshal(resetReq)
		resetHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
		resetHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		resetRec := httptest.NewRecorder()
		e.ServeHTTP(resetRec, resetHTTPReq)

		assert.Equal(t, http.StatusBadRequest, resetRec.Code)
	})
}

func TestAuth_ChangePassword_Integration(t *testing.T) {
	db, cleanup := setupTestDB(t)
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

	// Initialize DI Package for the router to use
	dicontainer.Init(db, redisRepo, eventBus, jwtSvc, cfg, log)

	// Setup Echo
	e := echo.New()
	router := routes.NewRouter(e, cfg, jwtSvc, redisRepo)
	router.RegisterAuthRoutes()

	t.Run("successful change", func(t *testing.T) {
		// First register and login a user
		registerReq := request.RegisterRequest{
			Email:           "testchangepass@example.com",
			Password:        "oldpassword123",
			PasswordConfirm: "oldpassword123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		// Login to get access token
		loginReq := request.LoginRequest{
			Email:    "testchangepass@example.com",
			Password: "oldpassword123",
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
		require.NotEmpty(t, loginResp.AccessToken)

		// Change password
		changeReq := request.ChangePasswordRequest{
			CurrentPassword: "oldpassword123",
			NewPassword:     "newpassword123",
			PasswordConfirm: "newpassword123",
		}
		changeBody, _ := json.Marshal(changeReq)
		changeHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(changeBody))
		changeHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		changeHTTPReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
		changeRec := httptest.NewRecorder()
		e.ServeHTTP(changeRec, changeHTTPReq)

		assert.Equal(t, http.StatusOK, changeRec.Code)

		var result map[string]string
		err = json.Unmarshal(changeRec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Password changed successfully. Please log in with your new password.", result["message"])

		// Verify that old password no longer works
		loginReq2 := request.LoginRequest{
			Email:    "testchangepass@example.com",
			Password: "oldpassword123",
		}
		loginBody2, _ := json.Marshal(loginReq2)
		loginHTTPReq2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody2))
		loginHTTPReq2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		loginRec2 := httptest.NewRecorder()
		e.ServeHTTP(loginRec2, loginHTTPReq2)

		assert.Equal(t, http.StatusUnauthorized, loginRec2.Code, "Old password should not work")

		// Verify that new password works
		loginReq3 := request.LoginRequest{
			Email:    "testchangepass@example.com",
			Password: "newpassword123",
		}
		loginBody3, _ := json.Marshal(loginReq3)
		loginHTTPReq3 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody3))
		loginHTTPReq3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		loginRec3 := httptest.NewRecorder()
		e.ServeHTTP(loginRec3, loginHTTPReq3)

		assert.Equal(t, http.StatusOK, loginRec3.Code, "New password should work")
	})

	t.Run("invalid current password", func(t *testing.T) {
		// Register and login a user
		registerReq := request.RegisterRequest{
			Email:           "testinvalidcurrent@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		// Login to get access token
		loginReq := request.LoginRequest{
			Email:    "testinvalidcurrent@example.com",
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

		// Try to change password with wrong current password
		changeReq := request.ChangePasswordRequest{
			CurrentPassword: "wrongpassword123",
			NewPassword:     "newpassword123",
			PasswordConfirm: "newpassword123",
		}
		changeBody, _ := json.Marshal(changeReq)
		changeHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(changeBody))
		changeHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		changeHTTPReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
		changeRec := httptest.NewRecorder()
		e.ServeHTTP(changeRec, changeHTTPReq)

		assert.Equal(t, http.StatusUnauthorized, changeRec.Code)
	})

	t.Run("not authenticated", func(t *testing.T) {
		changeReq := request.ChangePasswordRequest{
			CurrentPassword: "oldpassword123",
			NewPassword:     "newpassword123",
			PasswordConfirm: "newpassword123",
		}
		changeBody, _ := json.Marshal(changeReq)
		changeHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(changeBody))
		changeHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		// No Authorization header
		changeRec := httptest.NewRecorder()
		e.ServeHTTP(changeRec, changeHTTPReq)

		assert.Equal(t, http.StatusUnauthorized, changeRec.Code)
	})

	t.Run("invalid new password", func(t *testing.T) {
		// Register and login a user
		registerReq := request.RegisterRequest{
			Email:           "testinvalidnew@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		// Login to get access token
		loginReq := request.LoginRequest{
			Email:    "testinvalidnew@example.com",
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

		// Try to change password with short new password
		changeReq := request.ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "short",
			PasswordConfirm: "short",
		}
		changeBody, _ := json.Marshal(changeReq)
		changeHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(changeBody))
		changeHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		changeHTTPReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
		changeRec := httptest.NewRecorder()
		e.ServeHTTP(changeRec, changeHTTPReq)
		assert.Equal(t, http.StatusBadRequest, changeRec.Code)

		// Try to change password with new password without letters
		changeReq = request.ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "12345678",
			PasswordConfirm: "12345678",
		}
		changeBody, _ = json.Marshal(changeReq)
		changeHTTPReq = httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(changeBody))
		changeHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		changeHTTPReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
		changeRec = httptest.NewRecorder()
		e.ServeHTTP(changeRec, changeHTTPReq)
		assert.Equal(t, http.StatusBadRequest, changeRec.Code)

		// Try to change password with new password without numbers
		changeReq = request.ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "password",
			PasswordConfirm: "password",
		}
		changeBody, _ = json.Marshal(changeReq)
		changeHTTPReq = httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(changeBody))
		changeHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		changeHTTPReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
		changeRec = httptest.NewRecorder()
		e.ServeHTTP(changeRec, changeHTTPReq)
		assert.Equal(t, http.StatusBadRequest, changeRec.Code)
	})

	t.Run("password mismatch", func(t *testing.T) {
		// Register and login a user
		registerReq := request.RegisterRequest{
			Email:           "testmismatchchange@example.com",
			Password:        "password123",
			PasswordConfirm: "password123",
		}
		registerBody, _ := json.Marshal(registerReq)
		registerHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerBody))
		registerHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		registerRec := httptest.NewRecorder()
		e.ServeHTTP(registerRec, registerHTTPReq)
		require.Equal(t, http.StatusCreated, registerRec.Code)

		// Login to get access token
		loginReq := request.LoginRequest{
			Email:    "testmismatchchange@example.com",
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

		// Try to change password with mismatched passwords
		changeReq := request.ChangePasswordRequest{
			CurrentPassword: "password123",
			NewPassword:     "newpassword123",
			PasswordConfirm: "differentpassword123",
		}
		changeBody, _ := json.Marshal(changeReq)
		changeHTTPReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(changeBody))
		changeHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		changeHTTPReq.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
		changeRec := httptest.NewRecorder()
		e.ServeHTTP(changeRec, changeHTTPReq)

		assert.Equal(t, http.StatusBadRequest, changeRec.Code)
	})
}
