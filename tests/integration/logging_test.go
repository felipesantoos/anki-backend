package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// TestLoggingIntegration_CompleteFlow testa o fluxo completo de logging
// Handler → Middleware → Logger
func TestLoggingIntegration_CompleteFlow(t *testing.T) {
	// Setup: Create logger with buffer to capture logs
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Setup: Create test handler that uses the logger
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from context
		requestID := middlewares.GetRequestID(r.Context())
		if requestID == "" {
			t.Error("Request ID should be available in context")
		}

		// Use logger in handler
		testLogger.Info("Handler processing request",
			"request_id", requestID,
			"user_id", "123",
		)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply middleware
	middleware := middlewares.LoggingMiddlewareWithLogger(testLogger)
	wrappedHandler := middleware(testHandler)

	// Create request
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(`{"data":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Verifications
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify Request ID in header
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header")
	}

	// Verify generated logs
	logOutput := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 log lines (Request started, Handler log, Request completed), got %d", len(lines))
	}

	// Verify that Request ID appears in all logs
	for _, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse log line as JSON: %v, line: %s", err, line)
			continue
		}

		if logEntry["request_id"] != requestID {
			t.Errorf("Expected request_id %s in log, got %v", requestID, logEntry["request_id"])
		}
	}

	// Verify that specific logs were generated
	foundRequestStarted := false
	foundHandlerLog := false
	foundRequestCompleted := false

	for _, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			continue
		}

		msg, ok := logEntry["msg"].(string)
		if !ok {
			continue
		}

		switch msg {
		case "Request started":
			foundRequestStarted = true
			// Verify expected fields
			if logEntry["method"] != "POST" {
				t.Errorf("Expected method POST, got %v", logEntry["method"])
			}
			if logEntry["path"] != "/api/test" {
				t.Errorf("Expected path /api/test, got %v", logEntry["path"])
			}
		case "Handler processing request":
			foundHandlerLog = true
			// Verify that user_id is present
			if logEntry["user_id"] != "123" {
				t.Errorf("Expected user_id 123, got %v", logEntry["user_id"])
			}
		case "Request completed":
			foundRequestCompleted = true
			// Verify metrics
			if logEntry["status_code"] != float64(200) {
				t.Errorf("Expected status_code 200, got %v", logEntry["status_code"])
			}
			if logEntry["duration_ms"] == nil {
				t.Error("Expected duration_ms in log")
			}
			if logEntry["bytes_written"] == nil {
				t.Error("Expected bytes_written in log")
			}
		}
	}

	if !foundRequestStarted {
		t.Error("Expected 'Request started' log entry")
	}
	if !foundHandlerLog {
		t.Error("Expected 'Handler processing request' log entry")
	}
	if !foundRequestCompleted {
		t.Error("Expected 'Request completed' log entry")
	}
}

// TestLoggingIntegration_RequestIDPropagation testa que o Request ID
// is correctly propagated through the context
func TestLoggingIntegration_RequestIDPropagation(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Handler que verifica Request ID no contexto
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := middlewares.GetRequestID(ctx)

		if requestID == "" {
			t.Error("Request ID should be available in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Verify that it's the same Request ID as in the header
		headerRequestID := r.Header.Get("X-Request-ID")
		if headerRequestID != "" && headerRequestID != requestID {
			t.Errorf("Request ID mismatch: context=%s, header=%s", requestID, headerRequestID)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := middlewares.LoggingMiddlewareWithLogger(testLogger)
	wrappedHandler := middleware(testHandler)

	// Test 1: Request ID provided by client
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Request-ID", "custom-request-id-123")
	w1 := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w1, req1)

	if w1.Header().Get("X-Request-ID") != "custom-request-id-123" {
		t.Errorf("Expected custom request ID, got %s", w1.Header().Get("X-Request-ID"))
	}

	// Test 2: Request ID automatically generated
	logBuffer.Reset()
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w2, req2)

	generatedRequestID := w2.Header().Get("X-Request-ID")
	if generatedRequestID == "" {
		t.Error("Expected generated request ID")
	}
	if len(generatedRequestID) != 32 {
		t.Errorf("Expected request ID length 32, got %d", len(generatedRequestID))
	}

	// Verify that Request ID appears in logs
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, generatedRequestID) {
		t.Errorf("Expected request ID %s in logs, got: %s", generatedRequestID, logOutput)
	}
}

// TestLoggingIntegration_ErrorHandling testa logging de erros
func TestLoggingIntegration_ErrorHandling(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Handler que retorna erro
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	middleware := middlewares.LoggingMiddlewareWithLogger(testLogger)
	wrappedHandler := middleware(errorHandler)

	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Verify that error was logged
	logOutput := logBuffer.String()
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	
	// Get last line (Request completed)
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if err := json.Unmarshal([]byte(lastLine), &logEntry); err == nil {
			if logEntry["status_code"] != float64(404) {
				t.Errorf("Expected status_code 404 in log, got %v", logEntry["status_code"])
			}
		}
	}
}

// TestLoggingIntegration_PanicRecovery tests that panics are captured and logged
func TestLoggingIntegration_PanicRecovery(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Handler que causa panic
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := middlewares.LoggingMiddlewareWithLogger(testLogger)
	wrappedHandler := middleware(panicHandler)

	req := httptest.NewRequest("GET", "/api/panic", nil)
	w := httptest.NewRecorder()

	// Execute (should not cause panic in test)
	wrappedHandler.ServeHTTP(w, req)

	// Verify that status 500 was returned
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 after panic, got %d", w.Code)
	}

	// Verify that panic was logged
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "Request panicked") {
		t.Errorf("Expected 'Request panicked' in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "test panic") {
		t.Errorf("Expected panic message in logs, got: %s", logOutput)
	}
}

// TestLoggingIntegration_ContextLogger testa uso do logger com contexto
func TestLoggingIntegration_ContextLogger(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Handler que usa LogWithContext
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middlewares.GetRequestID(r.Context())
		
		ctx := map[string]interface{}{
			"request_id": requestID,
			"user_id":    "456",
			"action":     "create_deck",
		}

		contextLogger := logger.LogWithContext(testLogger, ctx)
		contextLogger.Info("Action completed")

		w.WriteHeader(http.StatusCreated)
	})

	middleware := middlewares.LoggingMiddlewareWithLogger(testLogger)
	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("POST", "/api/decks", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Verify that context was added to logs
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "user_id") {
		t.Errorf("Expected user_id in logs, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "action") {
		t.Errorf("Expected action in logs, got: %s", logOutput)
	}
}

