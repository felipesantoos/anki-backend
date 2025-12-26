package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/felipesantos/anki-backend/config"
)

func TestCORSMiddleware_Enabled(t *testing.T) {
	e := echo.New()
	e.Use(CORSMiddleware(config.CORSConfig{
		Enabled:         true,
		AllowedOrigins:  []string{"*"},
		AllowCredentials: true,
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify CORS headers are present
	if rec.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set")
	}
}

func TestCORSMiddleware_Disabled(t *testing.T) {
	e := echo.New()
	e.Use(CORSMiddleware(config.CORSConfig{
		Enabled: false,
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify CORS headers are NOT present when disabled
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Expected Access-Control-Allow-Origin header to NOT be set when CORS is disabled")
	}
}

func TestCORSMiddleware_AllowedOrigins(t *testing.T) {
	e := echo.New()
	e.Use(CORSMiddleware(config.CORSConfig{
		Enabled:         true,
		AllowedOrigins:  []string{"http://localhost:3000", "https://app.example.com"},
		AllowCredentials: false,
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
	}{
		{
			name:           "Allowed origin",
			origin:         "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
		},
		{
			name:           "Another allowed origin",
			origin:         "https://app.example.com",
			expectedOrigin: "https://app.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			actualOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if actualOrigin != tt.expectedOrigin {
				t.Errorf("Expected Access-Control-Allow-Origin to be %s, got %s", tt.expectedOrigin, actualOrigin)
			}
		})
	}
}

func TestCORSMiddleware_Credentials(t *testing.T) {
	tests := []struct {
		name              string
		allowCredentials  bool
		expectedHeader    string
	}{
		{
			name:             "Credentials allowed",
			allowCredentials: true,
			expectedHeader:   "true",
		},
		{
			name:             "Credentials not allowed",
			allowCredentials: false,
			expectedHeader:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(CORSMiddleware(config.CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"http://localhost:3000"},
				AllowCredentials: tt.allowCredentials,
			}))

			e.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			actualHeader := rec.Header().Get("Access-Control-Allow-Credentials")
			if actualHeader != tt.expectedHeader {
				t.Errorf("Expected Access-Control-Allow-Credentials to be %s, got %s", tt.expectedHeader, actualHeader)
			}
		})
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	e := echo.New()
	e.Use(CORSMiddleware(config.CORSConfig{
		Enabled:          true,
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
	}))

	e.POST("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Verify preflight response headers
	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("Expected Access-Control-Allow-Origin header in preflight response")
	}

	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header in preflight response")
	}

	if rec.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Expected Access-Control-Allow-Headers header in preflight response")
	}

	if rec.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("Expected Access-Control-Max-Age header in preflight response")
	}
}

func TestCORSMiddleware_Methods(t *testing.T) {
	e := echo.New()
	e.Use(CORSMiddleware(config.CORSConfig{
		Enabled:          true,
		AllowedOrigins:   []string{"*"},
		AllowCredentials: false,
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("Expected Access-Control-Allow-Methods header to be set")
	}

	// Verify common methods are included
	methodsToCheck := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	for _, method := range methodsToCheck {
		if !containsMethod(methods, method) {
			t.Errorf("Expected Access-Control-Allow-Methods to include %s, got %s", method, methods)
		}
	}
}

func TestCORSMiddleware_Headers(t *testing.T) {
	e := echo.New()
	e.Use(CORSMiddleware(config.CORSConfig{
		Enabled:          true,
		AllowedOrigins:   []string{"*"},
		AllowCredentials: false,
	}))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	headers := rec.Header().Get("Access-Control-Allow-Headers")
	if headers == "" {
		t.Error("Expected Access-Control-Allow-Headers header to be set")
	}

	// Verify common headers are included
	headersToCheck := []string{"Content-Type", "Authorization", "X-Request-ID"}
	for _, header := range headersToCheck {
		if !containsHeader(headers, header) {
			t.Errorf("Expected Access-Control-Allow-Headers to include %s, got %s", header, headers)
		}
	}
}

// Helper function to check if a method is in the methods string (case-insensitive)
func containsMethod(methods, method string) bool {
	return strings.Contains(strings.ToUpper(methods), strings.ToUpper(method))
}

// Helper function to check if a header is in the headers string (case-insensitive)
func containsHeader(headers, header string) bool {
	if len(headers) == 0 {
		return false
	}
	// Convert to uppercase for case-insensitive comparison
	upperHeaders := strings.ToUpper(headers)
	upperHeader := strings.ToUpper(header)
	return strings.Contains(upperHeaders, upperHeader)
}
