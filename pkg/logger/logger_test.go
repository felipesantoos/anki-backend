package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestSetupLogger_Development(t *testing.T) {
	var buf bytes.Buffer
	logger := SetupLogger(LoggerConfig{
		Level:       "INFO",
		Environment: "development",
		Output:      &buf,
	})

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "level=INFO") {
		t.Errorf("Expected text format, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestSetupLogger_Production(t *testing.T) {
	var buf bytes.Buffer
	logger := SetupLogger(LoggerConfig{
		Level:       "INFO",
		Environment: "production",
		Output:      &buf,
	})

	logger.Info("test message", "key", "value")

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Expected JSON format, got error: %v, output: %s", err, output)
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got: %v", logEntry["level"])
	}
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected message 'test message', got: %v", logEntry["msg"])
	}
}

func TestSetupLogger_Levels(t *testing.T) {
	tests := []struct {
		levelStr string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"INVALID", slog.LevelInfo}, // default
	}

	for _, tt := range tests {
		t.Run(tt.levelStr, func(t *testing.T) {
			var buf bytes.Buffer
			logger := SetupLogger(LoggerConfig{
				Level:       tt.levelStr,
				Environment: "development",
				Output:      &buf,
			})

			// Verify that the logger was created
			if logger == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}

func TestInitLogger(t *testing.T) {
	InitLogger("DEBUG", "development")
	logger := GetLogger()

	if logger == nil {
		t.Error("Logger should not be nil after InitLogger")
	}
}

func TestGetLogger_WithoutInit(t *testing.T) {
	// Reset global logger
	defaultLogger = nil

	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger should return a logger even without InitLogger")
	}
}

func TestLogWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := SetupLogger(LoggerConfig{
		Level:       "INFO",
		Environment: "development",
		Output:      &buf,
	})

	ctx := map[string]interface{}{
		"user_id":  "123",
		"request_id": "abc",
	}

	contextLogger := LogWithContext(logger, ctx)
	contextLogger.Info("test with context")

	output := buf.String()
	if !strings.Contains(output, "user_id=123") {
		t.Errorf("Expected user_id in output, got: %s", output)
	}
	if !strings.Contains(output, "request_id=abc") {
		t.Errorf("Expected request_id in output, got: %s", output)
	}
}

