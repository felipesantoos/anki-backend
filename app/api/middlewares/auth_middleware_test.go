package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// Ensure mockCacheRepository implements secondary.ICacheRepository
var _ secondary.ICacheRepository = (*mockCacheRepository)(nil)

// mockCacheRepository is a mock implementation of ICacheRepository
type mockCacheRepository struct {
	existsFunc func(ctx context.Context, key string) (bool, error)
}

func (m *mockCacheRepository) Ping(ctx context.Context) error {
	return nil
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, key)
	}
	return false, nil
}

func (m *mockCacheRepository) SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return false, nil
}

func (m *mockCacheRepository) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (m *mockCacheRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func createTestJWTService(t *testing.T) *jwt.JWTService {
	cfg := config.JWTConfig{
		SecretKey:          "test-secret-key-must-be-at-least-32-characters-long",
		AccessTokenExpiry:  15,
		RefreshTokenExpiry: 7,
		Issuer:             "test",
	}
	jwtSvc, err := jwt.NewJWTService(cfg)
	require.NoError(t, err)
	return jwtSvc
}

func TestAuthMiddleware_Success(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{
		existsFunc: func(ctx context.Context, key string) (bool, error) {
			return false, nil // Token not blacklisted
		},
	}

	middleware := AuthMiddleware(jwtSvc, cacheRepo)

	// Generate a valid access token
	userID := int64(123)
	token, err := jwtSvc.GenerateAccessToken(userID)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := middleware(func(c echo.Context) error {
		handlerCalled = true
		// Verify userID is set in context
		extractedUserID := GetUserID(c)
		assert.Equal(t, userID, extractedUserID)
		return c.String(http.StatusOK, "OK")
	})

	err = handler(c)
	require.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}

	middleware := AuthMiddleware(jwtSvc, cacheRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Contains(t, httpErr.Message.(string), "Authorization header is required")
	assert.False(t, handlerCalled)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}

	middleware := AuthMiddleware(jwtSvc, cacheRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Contains(t, httpErr.Message.(string), "Invalid authorization header format")
	assert.False(t, handlerCalled)
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{}

	middleware := AuthMiddleware(jwtSvc, cacheRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Contains(t, httpErr.Message.(string), "Token is required")
	assert.False(t, handlerCalled)
}

func TestAuthMiddleware_BlacklistedToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	token, _ := jwtSvc.GenerateAccessToken(123)
	
	cacheRepo := &mockCacheRepository{
		existsFunc: func(ctx context.Context, key string) (bool, error) {
			// Token is blacklisted
			return true, nil
		},
	}

	middleware := AuthMiddleware(jwtSvc, cacheRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Contains(t, httpErr.Message.(string), "Token has been invalidated")
	assert.False(t, handlerCalled)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	cacheRepo := &mockCacheRepository{
		existsFunc: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
	}

	middleware := AuthMiddleware(jwtSvc, cacheRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Contains(t, httpErr.Message.(string), "Invalid or expired token")
	assert.False(t, handlerCalled)
}

func TestAuthMiddleware_WrongTokenType(t *testing.T) {
	jwtSvc := createTestJWTService(t)
	// Generate refresh token instead of access token
	refreshToken, err := jwtSvc.GenerateRefreshToken(123)
	require.NoError(t, err)

	cacheRepo := &mockCacheRepository{
		existsFunc: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
	}

	middleware := AuthMiddleware(jwtSvc, cacheRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := middleware(func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	})

	err = handler(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Contains(t, httpErr.Message.(string), "Invalid token type")
	assert.False(t, handlerCalled)
}

func TestGetUserID(t *testing.T) {
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/test", nil), httptest.NewRecorder())

	// Initially, userID should be 0
	userID := GetUserID(c)
	assert.Equal(t, int64(0), userID)

	// Set userID in context
	c.Set(UserIDContextKey, int64(123))
	userID = GetUserID(c)
	assert.Equal(t, int64(123), userID)
}

func TestGetAccessToken(t *testing.T) {
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/test", nil), httptest.NewRecorder())

	// Initially, token should be empty
	token := GetAccessToken(c)
	assert.Equal(t, "", token)

	// Set token in context
	c.Set(AccessTokenContextKey, "test-token")
	token = GetAccessToken(c)
	assert.Equal(t, "test-token", token)
}

