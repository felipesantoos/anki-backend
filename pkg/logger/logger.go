package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// LoggerConfig contains logger configuration
type LoggerConfig struct {
	Level       string // "DEBUG", "INFO", "WARN", "ERROR"
	Environment string // "development", "staging", "production"
	Output      io.Writer
}

// SetupLogger configures structured logger based on configuration
func SetupLogger(config LoggerConfig) *slog.Logger {
	var level slog.Level
	switch strings.ToUpper(config.Level) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler

	// Use JSON handler for production/staging, Text handler for development
	if config.Environment == "production" || config.Environment == "staging" {
		// JSON formatter for production (easy parsing by aggregators)
		handler = slog.NewJSONHandler(config.Output, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})
	} else {
		// Simple formatter for development
		handler = slog.NewTextHandler(config.Output, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})
	}

	return slog.New(handler)
}

// Global logger (singleton)
var defaultLogger *slog.Logger

// InitLogger initializes global logger
func InitLogger(level, environment string) {
	defaultLogger = SetupLogger(LoggerConfig{
		Level:       level,
		Environment: environment,
		Output:      os.Stdout,
	})
}

// GetLogger returns global logger
// If not initialized, creates one with default values
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		InitLogger("INFO", "development")
	}
	return defaultLogger
}

// LogWithContext adds context to the logger
// Returns a new logger with the context fields added
func LogWithContext(logger *slog.Logger, ctx map[string]interface{}) *slog.Logger {
	args := make([]interface{}, 0, len(ctx)*2)
	for k, v := range ctx {
		args = append(args, k, v)
	}
	return logger.With(args...)
}

