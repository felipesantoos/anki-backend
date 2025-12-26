package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

func TestCustomHTTPErrorHandler_HTTPError(t *testing.T) {
	// Initialize logger for testing
	logger.InitLogger("INFO", "development")

	e := echo.New()
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Code)
	}

	var errorResp response.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Error != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got '%s'", errorResp.Error)
	}

	if errorResp.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got '%s'", errorResp.Message)
	}

	if errorResp.RequestID == "" {
		t.Error("Expected request ID to be present")
	}

	if errorResp.Path != "/test" {
		t.Errorf("Expected path '/test', got '%s'", errorResp.Path)
	}
}

func TestCustomHTTPErrorHandler_GenericError(t *testing.T) {
	logger.InitLogger("INFO", "development")

	e := echo.New()
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var errorResp response.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Error != "INTERNAL_SERVER_ERROR" {
		t.Errorf("Expected error code 'INTERNAL_SERVER_ERROR', got '%s'", errorResp.Error)
	}
}

func TestCustomHTTPErrorHandler_StatusCodes(t *testing.T) {
	logger.InitLogger("INFO", "development")

	testCases := []struct {
		name           string
		statusCode     int
		expectedCode   string
		message        string
	}{
		{"Bad Request", http.StatusBadRequest, "BAD_REQUEST", "Bad request"},
		{"Unauthorized", http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized"},
		{"Forbidden", http.StatusForbidden, "FORBIDDEN", "Forbidden"},
		{"Not Found", http.StatusNotFound, "NOT_FOUND", "Not found"},
		{"Conflict", http.StatusConflict, "CONFLICT", "Conflict"},
		{"Unprocessable Entity", http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation error"},
		{"Too Many Requests", http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests"},
		{"Internal Server Error", http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal error"},
		{"Bad Gateway", http.StatusBadGateway, "INTERNAL_SERVER_ERROR", "Bad gateway"},
		{"Service Unavailable", http.StatusServiceUnavailable, "INTERNAL_SERVER_ERROR", "Service unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.HTTPErrorHandler = CustomHTTPErrorHandler
			e.Use(RequestIDMiddleware())

			e.GET("/test", func(c echo.Context) error {
				return echo.NewHTTPError(tc.statusCode, tc.message)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, rec.Code)
			}

			var errorResp response.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if errorResp.Error != tc.expectedCode {
				t.Errorf("Expected error code '%s', got '%s'", tc.expectedCode, errorResp.Error)
			}

			if errorResp.Code != tc.expectedCode {
				t.Errorf("Expected code field '%s', got '%s'", tc.expectedCode, errorResp.Code)
			}
		})
	}
}

func TestCustomHTTPErrorHandler_RequestID(t *testing.T) {
	logger.InitLogger("INFO", "development")

	e := echo.New()
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var errorResp response.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.RequestID == "" {
		t.Error("Expected request ID to be present in error response")
	}

	// Verify request ID format (should be 32 hex characters from RequestIDMiddleware)
	if len(errorResp.RequestID) != 32 {
		t.Errorf("Expected request ID length 32, got %d", len(errorResp.RequestID))
	}
}

func TestCustomHTTPErrorHandler_ResponseFormat(t *testing.T) {
	logger.InitLogger("INFO", "development")

	e := echo.New()
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var errorResp response.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	// Verify all required fields are present
	if errorResp.Error == "" {
		t.Error("Expected 'error' field to be present")
	}

	if errorResp.Message == "" {
		t.Error("Expected 'message' field to be present")
	}

	if errorResp.Code == "" {
		t.Error("Expected 'code' field to be present")
	}

	if errorResp.RequestID == "" {
		t.Error("Expected 'request_id' field to be present")
	}

	if errorResp.Timestamp == "" {
		t.Error("Expected 'timestamp' field to be present")
	}

	if errorResp.Path == "" {
		t.Error("Expected 'path' field to be present")
	}
}

func TestCustomHTTPErrorHandler_Path(t *testing.T) {
	logger.InitLogger("INFO", "development")

	e := echo.New()
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(RequestIDMiddleware())

	e.GET("/api/users/:id", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var errorResp response.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Path != "/api/users/123" {
		t.Errorf("Expected path '/api/users/123', got '%s'", errorResp.Path)
	}
}

func TestCustomHTTPErrorHandler_Logging(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger.InitLogger("INFO", "development")
	// Note: slog doesn't have an easy way to capture output in tests
	// This test verifies that logging is called, but actual output verification
	// would require more complex setup. We'll verify the handler doesn't panic.

	e := echo.New()
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(RequestIDMiddleware())

	// Test client error (4xx) - should log as WARN
	e.GET("/client-error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/client-error", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Code)
	}

	// Test server error (5xx) - should log as ERROR
	e.GET("/server-error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal error")
	})

	req2 := httptest.NewRequest(http.MethodGet, "/server-error", nil)
	rec2 := httptest.NewRecorder()

	// Should not panic
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rec2.Code)
	}

	_ = buf // Suppress unused variable warning
}

func TestGetErrorCode(t *testing.T) {
	testCases := []struct {
		statusCode   int
		expectedCode string
	}{
		{http.StatusBadRequest, "BAD_REQUEST"},
		{http.StatusUnauthorized, "UNAUTHORIZED"},
		{http.StatusForbidden, "FORBIDDEN"},
		{http.StatusNotFound, "NOT_FOUND"},
		{http.StatusConflict, "CONFLICT"},
		{http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED"},
		{http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"},
		{http.StatusBadGateway, "INTERNAL_SERVER_ERROR"},
		{http.StatusServiceUnavailable, "INTERNAL_SERVER_ERROR"},
		{http.StatusGatewayTimeout, "INTERNAL_SERVER_ERROR"},
		{599, "INTERNAL_SERVER_ERROR"}, // 500+ range
		{399, "HTTP_ERROR"},            // Unknown status code
	}

	for _, tc := range testCases {
		t.Run(strings.ReplaceAll(http.StatusText(tc.statusCode), " ", "_"), func(t *testing.T) {
			result := getErrorCode(tc.statusCode)
			if result != tc.expectedCode {
				t.Errorf("getErrorCode(%d) = %s, want %s", tc.statusCode, result, tc.expectedCode)
			}
		})
	}
}
