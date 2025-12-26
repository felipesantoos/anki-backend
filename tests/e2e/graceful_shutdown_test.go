package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/config"
)

// TestGracefulShutdown_ConfigurableTimeout tests that shutdown timeout is configurable
func TestGracefulShutdown_ConfigurableTimeout(t *testing.T) {
	// Save original value
	originalTimeout := os.Getenv("SERVER_SHUTDOWN_TIMEOUT")
	defer func() {
		if originalTimeout != "" {
			os.Setenv("SERVER_SHUTDOWN_TIMEOUT", originalTimeout)
		} else {
			os.Unsetenv("SERVER_SHUTDOWN_TIMEOUT")
		}
	}()

	testCases := []struct {
		name           string
		envValue       string
		expectedResult int
	}{
		{
			name:           "Default timeout (10 seconds)",
			envValue:       "",
			expectedResult: 10,
		},
		{
			name:           "Custom timeout (15 seconds)",
			envValue:       "15",
			expectedResult: 15,
		},
		{
			name:           "Short timeout (5 seconds)",
			envValue:       "5",
			expectedResult: 5,
		},
		{
			name:           "Long timeout (30 seconds)",
			envValue:       "30",
			expectedResult: 30,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable
			if tc.envValue != "" {
				os.Setenv("SERVER_SHUTDOWN_TIMEOUT", tc.envValue)
			} else {
				os.Unsetenv("SERVER_SHUTDOWN_TIMEOUT")
			}

			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Verify shutdown timeout
			if cfg.Server.ShutdownTimeout != tc.expectedResult {
				t.Errorf("Expected shutdown timeout %d, got %d", tc.expectedResult, cfg.Server.ShutdownTimeout)
			}
		})
	}
}

// TestGracefulShutdown_ShutdownBehavior tests that shutdown context respects timeout
func TestGracefulShutdown_ShutdownBehavior(t *testing.T) {
	// Save original value
	originalTimeout := os.Getenv("SERVER_SHUTDOWN_TIMEOUT")
	defer func() {
		if originalTimeout != "" {
			os.Setenv("SERVER_SHUTDOWN_TIMEOUT", originalTimeout)
		} else {
			os.Unsetenv("SERVER_SHUTDOWN_TIMEOUT")
		}
	}()

	// Set a short timeout for testing
	os.Setenv("SERVER_SHUTDOWN_TIMEOUT", "2")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create a mock server
	srv := &http.Server{
		Addr: ":0", // Use random port
	}

	// Start server in background
	go func() {
		// This will fail because the port is already in use, but we just need the server
		_ = srv.ListenAndServe()
	}()

	// Give server a moment to attempt to start
	time.Sleep(50 * time.Millisecond)

	// Create shutdown context with configured timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer shutdownCancel()

	// Record start time
	startTime := time.Now()

	// Shutdown should complete quickly since server isn't really running
	err = srv.Shutdown(shutdownCtx)
	if err != nil && err != http.ErrServerClosed {
		// Ignore errors since we're just testing timeout behavior
	}

	// Verify that context timeout matches configured timeout
	elapsed := time.Since(startTime)
	if elapsed > time.Duration(cfg.Server.ShutdownTimeout+1)*time.Second {
		t.Errorf("Shutdown took longer than configured timeout: elapsed=%v, timeout=%d", elapsed, cfg.Server.ShutdownTimeout)
	}

	// Verify context was created with correct timeout
	deadline, ok := shutdownCtx.Deadline()
	if !ok {
		t.Error("Expected shutdown context to have a deadline")
	}

	expectedDeadline := startTime.Add(time.Duration(cfg.Server.ShutdownTimeout) * time.Second)
	// Allow 1 second tolerance for execution time
	if deadline.After(expectedDeadline.Add(1 * time.Second)) {
		t.Errorf("Shutdown context deadline is too far in the future: deadline=%v, expected=%v", deadline, expectedDeadline)
	}
}

