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
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	authService "github.com/felipesantos/anki-backend/core/services/auth"
)

// mockAuthService is a mock implementation of IAuthService
type mockAuthService struct {
	registerFunc func(ctx context.Context, email string, password string) (*entities.User, error)
}

func (m *mockAuthService) Register(ctx context.Context, email string, password string) (*entities.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, email, password)
	}
	return nil, nil
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
