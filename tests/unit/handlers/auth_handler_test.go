package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/core/domain/entities"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
)

// mockAuthService is a mock implementation of IAuthService
type mockAuthService struct {
	registerFunc     func(ctx context.Context, email string, password string) (*entities.User, error)
	loginFunc        func(ctx context.Context, email string, password string) (*response.LoginResponse, error)
	refreshTokenFunc func(ctx context.Context, refreshToken string) (*response.TokenResponse, error)
	logoutFunc       func(ctx context.Context, refreshToken string) error
}

func (m *mockAuthService) Register(ctx context.Context, email string, password string) (*entities.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, email, password)
	}
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, email string, password string) (*response.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, email, password)
	}
	return nil, nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*response.TokenResponse, error) {
	if m.refreshTokenFunc != nil {
		return m.refreshTokenFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, refreshToken)
	}
	return nil
}

func createTestUser() *entities.User {
	email, _ := valueobjects.NewEmail("user@example.com")
	password, _ := valueobjects.NewPassword("password123")
	return &entities.User{
		ID:            1,
		Email:         email,
		PasswordHash:  password,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	testUser := createTestUser()
	mockService := &mockAuthService{
		registerFunc: func(ctx context.Context, email string, password string) (*entities.User, error) {
			return testUser, nil
		},
	}

	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"email":            "user@example.com",
		"password":         "password123",
		"password_confirm": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Register() status code = %d, want %d", rec.Code, http.StatusCreated)
	}

	var result response.RegisterResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("Register() failed to unmarshal response: %v", err)
	}

	if result.User.ID != testUser.ID {
		t.Errorf("Register() user.ID = %d, want %d", result.User.ID, testUser.ID)
	}

	if result.User.Email != testUser.Email.Value() {
		t.Errorf("Register() user.Email = %v, want %v", result.User.Email, testUser.Email.Value())
	}
}

func TestAuthHandler_Register_InvalidRequest(t *testing.T) {
	mockService := &mockAuthService{}
	handler := handlers.NewAuthHandler(mockService)

	// Test with invalid JSON
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)

	if err == nil {
		t.Fatalf("Register() expected error for invalid JSON, got nil")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusBadRequest {
			t.Errorf("Register() status code = %d, want %d", httpErr.Code, http.StatusBadRequest)
		}
	} else {
		t.Errorf("Register() error type = %T, want *echo.HTTPError", err)
	}
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	mockService := &mockAuthService{
		registerFunc: func(ctx context.Context, email string, password string) (*entities.User, error) {
			return nil, authService.ErrEmailAlreadyExists
		},
	}

	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"email":            "existing@example.com",
		"password":         "password123",
		"password_confirm": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)

	if err == nil {
		t.Fatalf("Register() expected error, got nil")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusConflict {
			t.Errorf("Register() status code = %d, want %d", httpErr.Code, http.StatusConflict)
		}
	} else {
		t.Errorf("Register() error type = %T, want *echo.HTTPError", err)
	}
}

func TestAuthHandler_Register_ValidationErrors(t *testing.T) {
	mockService := &mockAuthService{}
	handler := handlers.NewAuthHandler(mockService)

	e := echo.New()

	tests := []struct {
		name    string
		reqBody map[string]interface{}
		wantCode int
	}{
		{
			name: "empty email",
			reqBody: map[string]interface{}{
				"email":            "",
				"password":         "password123",
				"password_confirm": "password123",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "empty password",
			reqBody: map[string]interface{}{
				"email":            "user@example.com",
				"password":         "",
				"password_confirm": "",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "password too short",
			reqBody: map[string]interface{}{
				"email":            "user@example.com",
				"password":         "pass1",
				"password_confirm": "pass1",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "password mismatch",
			reqBody: map[string]interface{}{
				"email":            "user@example.com",
				"password":         "password123",
				"password_confirm": "password456",
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Register(c)

			if err == nil {
				t.Fatalf("Register() expected error, got nil")
			}

			if httpErr, ok := err.(*echo.HTTPError); ok {
				if httpErr.Code != tt.wantCode {
					t.Errorf("Register() status code = %d, want %d", httpErr.Code, tt.wantCode)
				}
			} else {
				t.Errorf("Register() error type = %T, want *echo.HTTPError", err)
			}
		})
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	loginResp := &response.LoginResponse{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    900,
		TokenType:    "Bearer",
		User: response.UserData{
			ID:            1,
			Email:         "user@example.com",
			EmailVerified: false,
			CreatedAt:     time.Now(),
		},
	}

	mockService := &mockAuthService{
		loginFunc: func(ctx context.Context, email string, password string) (*response.LoginResponse, error) {
			return loginResp, nil
		},
	}

	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"email":    "user@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)

	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Login() status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var result response.LoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("Login() failed to unmarshal response: %v", err)
	}

	if result.AccessToken != loginResp.AccessToken {
		t.Errorf("Login() AccessToken = %v, want %v", result.AccessToken, loginResp.AccessToken)
	}

	if result.RefreshToken != loginResp.RefreshToken {
		t.Errorf("Login() RefreshToken = %v, want %v", result.RefreshToken, loginResp.RefreshToken)
	}

	if result.TokenType != "Bearer" {
		t.Errorf("Login() TokenType = %v, want 'Bearer'", result.TokenType)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := &mockAuthService{
		loginFunc: func(ctx context.Context, email string, password string) (*response.LoginResponse, error) {
			return nil, authService.ErrInvalidCredentials
		},
	}

	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"email":    "user@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)

	if err == nil {
		t.Fatalf("Login() expected error, got nil")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("Login() status code = %d, want %d", httpErr.Code, http.StatusUnauthorized)
		}
	} else {
		t.Errorf("Login() error type = %T, want *echo.HTTPError", err)
	}
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	mockService := &mockAuthService{}
	handler := handlers.NewAuthHandler(mockService)

	tests := []struct {
		name    string
		reqBody map[string]interface{}
		wantCode int
	}{
		{
			name: "empty email",
			reqBody: map[string]interface{}{
				"email":    "",
				"password": "password123",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "empty password",
			reqBody: map[string]interface{}{
				"email":    "user@example.com",
				"password": "",
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.reqBody)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Login(c)

			if err == nil {
				t.Fatalf("Login() expected error, got nil")
			}

			if httpErr, ok := err.(*echo.HTTPError); ok {
				if httpErr.Code != tt.wantCode {
					t.Errorf("Login() status code = %d, want %d", httpErr.Code, tt.wantCode)
				}
			} else {
				t.Errorf("Login() error type = %T, want *echo.HTTPError", err)
			}
		})
	}
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	tokenResp := &response.TokenResponse{
		AccessToken: "new-access-token",
		ExpiresIn:   900,
		TokenType:   "Bearer",
	}

	mockService := &mockAuthService{
		refreshTokenFunc: func(ctx context.Context, refreshToken string) (*response.TokenResponse, error) {
			return tokenResp, nil
		},
	}

	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"refresh_token": "valid-refresh-token",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.RefreshToken(c)

	if err != nil {
		t.Fatalf("RefreshToken() error = %v, want nil", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("RefreshToken() status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var result response.TokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("RefreshToken() failed to unmarshal response: %v", err)
	}

	if result.AccessToken != tokenResp.AccessToken {
		t.Errorf("RefreshToken() AccessToken = %v, want %v", result.AccessToken, tokenResp.AccessToken)
	}

	if result.TokenType != "Bearer" {
		t.Errorf("RefreshToken() TokenType = %v, want 'Bearer'", result.TokenType)
	}
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	mockService := &mockAuthService{
		refreshTokenFunc: func(ctx context.Context, refreshToken string) (*response.TokenResponse, error) {
			return nil, authService.ErrInvalidToken
		},
	}

	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"refresh_token": "invalid-token",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.RefreshToken(c)

	if err == nil {
		t.Fatalf("RefreshToken() expected error, got nil")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("RefreshToken() status code = %d, want %d", httpErr.Code, http.StatusUnauthorized)
		}
	} else {
		t.Errorf("RefreshToken() error type = %T, want *echo.HTTPError", err)
	}
}

func TestAuthHandler_RefreshToken_EmptyToken(t *testing.T) {
	mockService := &mockAuthService{}
	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"refresh_token": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.RefreshToken(c)

	if err == nil {
		t.Fatalf("RefreshToken() expected error, got nil")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusBadRequest {
			t.Errorf("RefreshToken() status code = %d, want %d", httpErr.Code, http.StatusBadRequest)
		}
	} else {
		t.Errorf("RefreshToken() error type = %T, want *echo.HTTPError", err)
	}
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockService := &mockAuthService{
		logoutFunc: func(ctx context.Context, refreshToken string) error {
			return nil
		},
	}

	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"refresh_token": "valid-refresh-token",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Logout(c)

	if err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Logout() status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("Logout() failed to unmarshal response: %v", err)
	}

	if result["message"] != "Logged out successfully" {
		t.Errorf("Logout() message = %v, want 'Logged out successfully'", result["message"])
	}
}

func TestAuthHandler_Logout_EmptyToken(t *testing.T) {
	mockService := &mockAuthService{}
	handler := handlers.NewAuthHandler(mockService)

	reqBody := map[string]interface{}{
		"refresh_token": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Logout(c)

	if err == nil {
		t.Fatalf("Logout() expected error, got nil")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusBadRequest {
			t.Errorf("Logout() status code = %d, want %d", httpErr.Code, http.StatusBadRequest)
		}
	} else {
		t.Errorf("Logout() error type = %T, want *echo.HTTPError", err)
	}
}
