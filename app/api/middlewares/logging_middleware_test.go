package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Handler de teste
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply middleware
	middleware := LoggingMiddlewareWithLogger(logger)
	wrappedHandler := middleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify that request ID was added to header
	if w.Header().Get("X-Request-ID") != "test-request-id" {
		t.Errorf("Expected X-Request-ID header, got: %s", w.Header().Get("X-Request-ID"))
	}

	// Verify logs
	output := buf.String()
	if !strings.Contains(output, "Request started") {
		t.Errorf("Expected 'Request started' in logs, got: %s", output)
	}
	if !strings.Contains(output, "Request completed") {
		t.Errorf("Expected 'Request completed' in logs, got: %s", output)
	}
	if !strings.Contains(output, "test-request-id") {
		t.Errorf("Expected request ID in logs, got: %s", output)
	}
}

func TestLoggingMiddleware_GenerateRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	})

	middleware := LoggingMiddlewareWithLogger(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/test", nil)
	// Don't set X-Request-ID - should be generated
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Verify that a request ID was generated
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected request ID to be generated")
	}
	if len(requestID) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("Expected request ID to be 32 chars, got %d: %s", len(requestID), requestID)
	}
}

func TestGetRequestID(t *testing.T) {
	ctx := context.Background()
	
	// Sem request ID
	if id := GetRequestID(ctx); id != "" {
		t.Errorf("Expected empty request ID, got: %s", id)
	}

	// Com request ID
	ctx = context.WithValue(ctx, requestIDKey{}, "test-id")
	if id := GetRequestID(ctx); id != "test-id" {
		t.Errorf("Expected 'test-id', got: %s", id)
	}
}

func TestResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rw.statusCode)
	}

	// Test Write
	data := []byte("test data")
	n, err := rw.Write(data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if rw.bytesWritten != int64(len(data)) {
		t.Errorf("Expected bytesWritten to be %d, got %d", len(data), rw.bytesWritten)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedIP     string
	}{
		{
			name:       "X-Forwarded-For",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1"},
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "10.0.0.1"},
			expectedIP: "10.0.0.1",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "127.0.0.1:8080",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1", "X-Real-IP": "10.0.0.1"},
			expectedIP: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			ip := getClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestLoggingMiddleware_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	})

	middleware := LoggingMiddlewareWithLogger(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/api/test?foo=bar", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Verify JSON format
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 log lines, got %d", len(lines))
	}

	// Verify first line (Request started)
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
		t.Errorf("Expected JSON format, got error: %v, line: %s", err, lines[0])
	}

	if logEntry["msg"] != "Request started" {
		t.Errorf("Expected 'Request started', got: %v", logEntry["msg"])
	}
	if logEntry["method"] != "GET" {
		t.Errorf("Expected method 'GET', got: %v", logEntry["method"])
	}
	if logEntry["path"] != "/api/test" {
		t.Errorf("Expected path '/api/test', got: %v", logEntry["path"])
	}
}

