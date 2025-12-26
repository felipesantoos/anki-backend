package e2e

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// setupTestServer creates a test HTTP server with logging configured
func setupTestServer(env string, logLevel string) (*httptest.Server, *bytes.Buffer) {
	var logBuffer bytes.Buffer

	// Configure logger based on environment
	var testLogger *slog.Logger
	if env == "production" || env == "staging" {
		testLogger = slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
			Level:     parseLogLevel(logLevel),
			AddSource: true,
		}))
	} else {
		testLogger = slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
			Level:     parseLogLevel(logLevel),
			AddSource: true,
		}))
	}

	// Create simple router for tests
	mux := http.NewServeMux()

	// Health check route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Test route that uses logger
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		requestID := middlewares.GetRequestID(r.Context())
		testLogger.Info("Processing test request",
			"request_id", requestID,
			"method", r.Method,
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	})

	// Apply logging middleware
	handler := middlewares.LoggingMiddlewareWithLogger(testLogger)(mux)

	server := httptest.NewServer(handler)
	return server, &logBuffer
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// TestLoggingE2E_ProductionFormat tests that logs in production are JSON
func TestLoggingE2E_ProductionFormat(t *testing.T) {
	server, logBuffer := setupTestServer("production", "INFO")
	defer server.Close()

	// Make request
	resp, err := http.Get(server.URL + "/api/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify Request ID in header
	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header in response")
	}

	// Verify JSON format of logs
	logOutput := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Expected JSON format in production, got error: %v, line: %s", err, line)
		}

		// Verify required fields
		if logEntry["time"] == nil {
			t.Error("Expected 'time' field in JSON log")
		}
		if logEntry["level"] == nil {
			t.Error("Expected 'level' field in JSON log")
		}
		if logEntry["msg"] == nil {
			t.Error("Expected 'msg' field in JSON log")
		}
	}
}

// TestLoggingE2E_DevelopmentFormat tests that logs in development are readable text
func TestLoggingE2E_DevelopmentFormat(t *testing.T) {
	server, logBuffer := setupTestServer("development", "INFO")
	defer server.Close()

	// Make request
	resp, err := http.Get(server.URL + "/api/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify text format of logs
	logOutput := logBuffer.String()

	// In development, logs should be readable text (not JSON)
	if strings.HasPrefix(strings.TrimSpace(logOutput), "{") {
		t.Error("Expected text format in development, got JSON")
	}

	// Verify that it contains readable fields
	if !strings.Contains(logOutput, "level=") {
		t.Error("Expected 'level=' in text format")
	}
	if !strings.Contains(logOutput, "msg=") {
		t.Error("Expected 'msg=' in text format")
	}
}

// TestLoggingE2E_RequestIDPropagation tests Request ID propagation in full flow
func TestLoggingE2E_RequestIDPropagation(t *testing.T) {
	server, logBuffer := setupTestServer("development", "INFO")
	defer server.Close()

	// Create HTTP client
	client := &http.Client{}

	// Test 1: Custom Request ID
	req1, _ := http.NewRequest("GET", server.URL+"/api/test", nil)
	req1.Header.Set("X-Request-ID", "custom-e2e-test-id")

	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.Header.Get("X-Request-ID") != "custom-e2e-test-id" {
		t.Errorf("Expected custom request ID, got %s", resp1.Header.Get("X-Request-ID"))
	}

	// Verify that Request ID appears in logs
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "custom-e2e-test-id") {
		t.Errorf("Expected request ID in logs, got: %s", logOutput)
	}

	// Test 2: Generated Request ID
	logBuffer.Reset()
	req2, _ := http.NewRequest("GET", server.URL+"/health", nil)

	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp2.Body.Close()

	generatedRequestID := resp2.Header.Get("X-Request-ID")
	if generatedRequestID == "" {
		t.Error("Expected generated request ID")
	}

	logOutput = logBuffer.String()
	if !strings.Contains(logOutput, generatedRequestID) {
		t.Errorf("Expected generated request ID %s in logs, got: %s", generatedRequestID, logOutput)
	}
}

// TestLoggingE2E_CompleteRequestFlow tests complete request flow with logging
func TestLoggingE2E_CompleteRequestFlow(t *testing.T) {
	server, logBuffer := setupTestServer("production", "INFO")
	defer server.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL+"/api/test", nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Response verifications
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header")
	}

	// Verify generated logs
	logOutput := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")

	// Deve ter pelo menos 2 logs: Request started e Request completed
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 log entries, got %d", len(lines))
	}

	// Verify Request started
	var startedLog map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &startedLog); err != nil {
		t.Fatalf("Failed to parse Request started log: %v", err)
	}

	if startedLog["msg"] != "Request started" {
		t.Errorf("Expected 'Request started' message, got %v", startedLog["msg"])
	}
	if startedLog["request_id"] != requestID {
		t.Errorf("Request ID mismatch in Request started log")
	}

	// Verify Request completed
	var completedLog map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &completedLog); err != nil {
		t.Fatalf("Failed to parse Request completed log: %v", err)
	}

	if completedLog["msg"] != "Request completed" {
		t.Errorf("Expected 'Request completed' message, got %v", completedLog["msg"])
	}
	if completedLog["status_code"] != float64(200) {
		t.Errorf("Expected status_code 200, got %v", completedLog["status_code"])
	}
	if completedLog["duration_ms"] == nil {
		t.Error("Expected duration_ms in Request completed log")
	}
}

// TestLoggingE2E_ConfigIntegration tests integration with configuration system
func TestLoggingE2E_ConfigIntegration(t *testing.T) {
	// Salvar valores originais
	originalEnv := os.Getenv("ENV")
	originalLogLevel := os.Getenv("LOG_LEVEL")

	// Cleanup after test
	defer func() {
		if originalEnv != "" {
			os.Setenv("ENV", originalEnv)
		} else {
			os.Unsetenv("ENV")
		}
		if originalLogLevel != "" {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Test 1: Production configuration
	os.Setenv("ENV", "production")
	os.Setenv("LOG_LEVEL", "WARN")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Logger.Environment != "production" {
		t.Errorf("Expected environment production, got %s", cfg.Logger.Environment)
	}
	if cfg.Logger.Level != "WARN" {
		t.Errorf("Expected log level WARN, got %s", cfg.Logger.Level)
	}

	// Initialize logger with configuration
	logger.InitLogger(cfg.Logger.Level, cfg.Logger.Environment)
	log := logger.GetLogger()

	if log == nil {
		t.Error("Logger should not be nil")
	}

	// Test 2: Development configuration
	os.Setenv("ENV", "development")
	os.Setenv("LOG_LEVEL", "DEBUG")

	cfg, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	logger.InitLogger(cfg.Logger.Level, cfg.Logger.Environment)
	log = logger.GetLogger()

	if log == nil {
		t.Error("Logger should not be nil")
	}
}

// TestLoggingE2E_MultipleRequests tests multiple requests and verifies
// that each has its own Request ID
func TestLoggingE2E_MultipleRequests(t *testing.T) {
	server, logBuffer := setupTestServer("production", "INFO")
	defer server.Close()

	client := &http.Client{}
	requestIDs := make(map[string]bool)

	// Make 5 requests
	for i := 0; i < 5; i++ {
		logBuffer.Reset()
		req, _ := http.NewRequest("GET", server.URL+"/api/test", nil)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request %d: %v", i, err)
		}

		requestID := resp.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Errorf("Request %d: Expected X-Request-ID header", i)
		}

		// Verify that each Request ID is unique
		if requestIDs[requestID] {
			t.Errorf("Request %d: Duplicate request ID %s", i, requestID)
		}
		requestIDs[requestID] = true

		// Verify that Request ID appears in logs
		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, requestID) {
			t.Errorf("Request %d: Expected request ID %s in logs", i, requestID)
		}

		resp.Body.Close()
	}

	if len(requestIDs) != 5 {
		t.Errorf("Expected 5 unique request IDs, got %d", len(requestIDs))
	}
}

