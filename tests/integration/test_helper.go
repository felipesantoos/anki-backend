package integration

import (
	"path/filepath"
	"runtime"

	"github.com/felipesantos/anki-backend/config"
)

// init loads .env.test file if it exists before running any integration tests.
// This ensures that test-specific environment variables (like DB_PORT=5433) are loaded
// automatically without requiring manual export before running tests.
func init() {
	// Try to load .env.test file from the project root
	// Get the path relative to this file: tests/integration/test_helper.go -> ../../.env.test
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	projectRoot := filepath.Join(testDir, "..", "..")
	envTestPath := filepath.Join(projectRoot, ".env.test")

	// Try to load .env.test (silently ignore if it doesn't exist)
	_ = config.LoadFromFile(envTestPath)
}

