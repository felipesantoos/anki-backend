package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase debug", "debug", "debug"},
		{"uppercase DEBUG", "DEBUG", "debug"},
		{"lowercase info", "info", "info"},
		{"uppercase INFO", "INFO", "info"},
		{"lowercase warn", "warn", "warn"},
		{"uppercase WARN", "WARN", "warn"},
		{"warning", "warning", "warn"},
		{"lowercase error", "error", "error"},
		{"uppercase ERROR", "ERROR", "error"},
		{"invalid level", "invalid", "info"},
		{"empty string", "", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("validateLogLevel(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase development", "development", "development"},
		{"dev", "dev", "development"},
		{"uppercase DEV", "DEV", "development"},
		{"lowercase staging", "staging", "staging"},
		{"stage", "stage", "staging"},
		{"uppercase STAGING", "STAGING", "staging"},
		{"lowercase production", "production", "production"},
		{"prod", "prod", "production"},
		{"uppercase PROD", "PROD", "production"},
		{"PRODUCTION", "PRODUCTION", "production"},
		{"invalid env", "invalid", "development"},
		{"empty string", "", "development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateEnvironment(tt.input)
			if result != tt.expected {
				t.Errorf("validateEnvironment(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseCORSOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Wildcard",
			input:    "*",
			expected: []string{"*"},
		},
		{
			name:     "Single origin",
			input:    "http://localhost:3000",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "Multiple origins comma-separated",
			input:    "http://localhost:3000,https://app.example.com",
			expected: []string{"http://localhost:3000", "https://app.example.com"},
		},
		{
			name:     "Multiple origins with spaces",
			input:    "http://localhost:3000, https://app.example.com , https://api.example.com",
			expected: []string{"http://localhost:3000", "https://app.example.com", "https://api.example.com"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Only whitespace",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "Comma-separated with empty parts",
			input:    "http://localhost:3000,,https://app.example.com",
			expected: []string{"http://localhost:3000", "https://app.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCORSOrigins(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseCORSOrigins(%q) length = %d, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseCORSOrigins(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestLoad_LoggerConfig(t *testing.T) {
	// Save original values
	originalEnv := os.Getenv("ENV")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	originalJWT := os.Getenv("JWT_SECRET_KEY")
	originalCORS := os.Getenv("CORS_ALLOWED_ORIGINS")

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
		if originalJWT != "" {
			os.Setenv("JWT_SECRET_KEY", originalJWT)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
		if originalCORS != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalCORS)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
	}()

	// Test 1: Valid values for production (requires JWT_SECRET_KEY and CORS_ALLOWED_ORIGINS)
	os.Setenv("ENV", "production")
	os.Setenv("LOG_LEVEL", "WARN")
	os.Setenv("JWT_SECRET_KEY", strings.Repeat("a", 32)) // Valid JWT secret
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Logger.Environment != "production" {
		t.Errorf("Expected environment 'production', got %q", cfg.Logger.Environment)
	}
	if cfg.Logger.Level != "warn" {
		t.Errorf("Expected log level 'warn', got %q", cfg.Logger.Level)
	}

	// Test 2: Invalid values (should use defaults)
	os.Setenv("ENV", "invalid_env")
	os.Setenv("LOG_LEVEL", "invalid_level")
	os.Unsetenv("JWT_SECRET_KEY")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Logger.Environment != "development" {
		t.Errorf("Expected default environment 'development', got %q", cfg.Logger.Environment)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("Expected default log level 'info', got %q", cfg.Logger.Level)
	}

	// Test 3: Undefined values (should use defaults)
	os.Unsetenv("ENV")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("JWT_SECRET_KEY")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Logger.Environment != "development" {
		t.Errorf("Expected default environment 'development', got %q", cfg.Logger.Environment)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("Expected default log level 'info', got %q", cfg.Logger.Level)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary directory for test .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env.test")

	// Create a test .env file
	envContent := `TEST_VAR=test_value
ENV=staging
LOG_LEVEL=debug
`
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Test loading from file
	err = LoadFromFile(envFile)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Verify values were loaded
	if os.Getenv("TEST_VAR") != "test_value" {
		t.Errorf("Expected TEST_VAR=test_value, got %q", os.Getenv("TEST_VAR"))
	}
	if os.Getenv("ENV") != "staging" {
		t.Errorf("Expected ENV=staging, got %q", os.Getenv("ENV"))
	}

	// Cleanup
	os.Unsetenv("TEST_VAR")
	os.Unsetenv("ENV")
	os.Unsetenv("LOG_LEVEL")
}

func TestLoad_EnvFilePriority(t *testing.T) {
	// Create a temporary directory for test .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env.priority")

	// Create a test .env file with a value
	envContent := `SERVER_PORT=9999
`
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Set environment variable (should take priority)
	originalPort := os.Getenv("SERVER_PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("SERVER_PORT", originalPort)
		} else {
			os.Unsetenv("SERVER_PORT")
		}
	}()

	os.Setenv("SERVER_PORT", "8888")

	// Load from file
	err = LoadFromFile(envFile)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	// Environment variable should take priority
	if os.Getenv("SERVER_PORT") != "8888" {
		t.Errorf("Expected SERVER_PORT=8888 (from env), got %q", os.Getenv("SERVER_PORT"))
	}

	// Unset env var and verify .env value is used
	os.Unsetenv("SERVER_PORT")
	err = LoadFromFile(envFile)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	_, err = loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	// After unsetting env var, should use value from .env (but Load() doesn't reload)
	// This test mainly verifies that LoadFromFile works correctly
}

func TestValidate_Development(t *testing.T) {
	// Save original values
	originalEnv := os.Getenv("ENV")
	originalJWT := os.Getenv("JWT_SECRET_KEY")

	defer func() {
		if originalEnv != "" {
			os.Setenv("ENV", originalEnv)
		} else {
			os.Unsetenv("ENV")
		}
		if originalJWT != "" {
			os.Setenv("JWT_SECRET_KEY", originalJWT)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
	}()

	os.Setenv("ENV", "development")
	os.Unsetenv("JWT_SECRET_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Development should not require JWT_SECRET_KEY
	if cfg.Logger.Environment != "development" {
		t.Errorf("Expected environment 'development', got %q", cfg.Logger.Environment)
	}
}

func TestValidate_Production_MissingJWT(t *testing.T) {
	// Save original values
	originalEnv := os.Getenv("ENV")
	originalJWT := os.Getenv("JWT_SECRET_KEY")
	originalCORS := os.Getenv("CORS_ALLOWED_ORIGINS")

	defer func() {
		if originalEnv != "" {
			os.Setenv("ENV", originalEnv)
		} else {
			os.Unsetenv("ENV")
		}
		if originalJWT != "" {
			os.Setenv("JWT_SECRET_KEY", originalJWT)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
		if originalCORS != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalCORS)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
	}()

	os.Setenv("ENV", "production")
	os.Unsetenv("JWT_SECRET_KEY")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing JWT_SECRET_KEY in production, got nil")
	}

	requiredErr, ok := err.(*RequiredEnvError)
	if !ok {
		t.Fatalf("Expected RequiredEnvError, got %T: %v", err, err)
	}

	if !contains(requiredErr.Variables, "JWT_SECRET_KEY") {
		t.Errorf("Expected JWT_SECRET_KEY in missing variables, got %v", requiredErr.Variables)
	}
	if !contains(requiredErr.Variables, "CORS_ALLOWED_ORIGINS") {
		t.Errorf("Expected CORS_ALLOWED_ORIGINS in missing variables, got %v", requiredErr.Variables)
	}
	if requiredErr.Environment != "production" {
		t.Errorf("Expected environment 'production', got %q", requiredErr.Environment)
	}
}

func TestValidate_Production_CORSWildcardWithCredentials(t *testing.T) {
	// Save original values
	originalEnv := os.Getenv("ENV")
	originalJWT := os.Getenv("JWT_SECRET_KEY")
	originalCORS := os.Getenv("CORS_ALLOWED_ORIGINS")
	originalCORSEnabled := os.Getenv("CORS_ENABLED")
	originalCORSCreds := os.Getenv("CORS_ALLOW_CREDENTIALS")

	defer func() {
		if originalEnv != "" {
			os.Setenv("ENV", originalEnv)
		} else {
			os.Unsetenv("ENV")
		}
		if originalJWT != "" {
			os.Setenv("JWT_SECRET_KEY", originalJWT)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
		if originalCORS != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalCORS)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
		if originalCORSEnabled != "" {
			os.Setenv("CORS_ENABLED", originalCORSEnabled)
		} else {
			os.Unsetenv("CORS_ENABLED")
		}
		if originalCORSCreds != "" {
			os.Setenv("CORS_ALLOW_CREDENTIALS", originalCORSCreds)
		} else {
			os.Unsetenv("CORS_ALLOW_CREDENTIALS")
		}
	}()

	os.Setenv("ENV", "production")
	os.Setenv("JWT_SECRET_KEY", strings.Repeat("a", 32)) // Valid JWT secret
	os.Setenv("CORS_ALLOWED_ORIGINS", "*")
	os.Setenv("CORS_ALLOW_CREDENTIALS", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for wildcard CORS with credentials in production, got nil")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T: %v", err, err)
	}

	errorMsg := validationErr.Error()
	if !strings.Contains(errorMsg, "wildcard") || !strings.Contains(errorMsg, "production") {
		t.Errorf("Expected validation error about wildcard CORS in production, got: %v", validationErr)
	}
}

func TestValidate_S3Storage_MissingConfig(t *testing.T) {
	// Save original values
	originalStorageType := os.Getenv("STORAGE_TYPE")
	originalS3Bucket := os.Getenv("STORAGE_S3_BUCKET")
	originalS3Region := os.Getenv("STORAGE_S3_REGION")
	originalS3Key := os.Getenv("STORAGE_S3_KEY")
	originalS3Secret := os.Getenv("STORAGE_S3_SECRET")

	defer func() {
		if originalStorageType != "" {
			os.Setenv("STORAGE_TYPE", originalStorageType)
		} else {
			os.Unsetenv("STORAGE_TYPE")
		}
		if originalS3Bucket != "" {
			os.Setenv("STORAGE_S3_BUCKET", originalS3Bucket)
		} else {
			os.Unsetenv("STORAGE_S3_BUCKET")
		}
		if originalS3Region != "" {
			os.Setenv("STORAGE_S3_REGION", originalS3Region)
		} else {
			os.Unsetenv("STORAGE_S3_REGION")
		}
		if originalS3Key != "" {
			os.Setenv("STORAGE_S3_KEY", originalS3Key)
		} else {
			os.Unsetenv("STORAGE_S3_KEY")
		}
		if originalS3Secret != "" {
			os.Setenv("STORAGE_S3_SECRET", originalS3Secret)
		} else {
			os.Unsetenv("STORAGE_S3_SECRET")
		}
	}()

	os.Setenv("STORAGE_TYPE", "s3")
	os.Unsetenv("STORAGE_S3_BUCKET")
	os.Unsetenv("STORAGE_S3_REGION")
	os.Unsetenv("STORAGE_S3_KEY")
	os.Unsetenv("STORAGE_S3_SECRET")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing S3 configuration, got nil")
	}

	requiredErr, ok := err.(*RequiredEnvError)
	if !ok {
		t.Fatalf("Expected RequiredEnvError, got %T: %v", err, err)
	}

	requiredS3Vars := []string{"STORAGE_S3_BUCKET", "STORAGE_S3_REGION", "STORAGE_S3_KEY", "STORAGE_S3_SECRET"}
	for _, varName := range requiredS3Vars {
		if !contains(requiredErr.Variables, varName) {
			t.Errorf("Expected %s in missing variables, got %v", varName, requiredErr.Variables)
		}
	}
}

func TestValidate_S3Storage_ValidConfig(t *testing.T) {
	// Save original values
	originalStorageType := os.Getenv("STORAGE_TYPE")
	originalS3Bucket := os.Getenv("STORAGE_S3_BUCKET")
	originalS3Region := os.Getenv("STORAGE_S3_REGION")
	originalS3Key := os.Getenv("STORAGE_S3_KEY")
	originalS3Secret := os.Getenv("STORAGE_S3_SECRET")

	defer func() {
		if originalStorageType != "" {
			os.Setenv("STORAGE_TYPE", originalStorageType)
		} else {
			os.Unsetenv("STORAGE_TYPE")
		}
		if originalS3Bucket != "" {
			os.Setenv("STORAGE_S3_BUCKET", originalS3Bucket)
		} else {
			os.Unsetenv("STORAGE_S3_BUCKET")
		}
		if originalS3Region != "" {
			os.Setenv("STORAGE_S3_REGION", originalS3Region)
		} else {
			os.Unsetenv("STORAGE_S3_REGION")
		}
		if originalS3Key != "" {
			os.Setenv("STORAGE_S3_KEY", originalS3Key)
		} else {
			os.Unsetenv("STORAGE_S3_KEY")
		}
		if originalS3Secret != "" {
			os.Setenv("STORAGE_S3_SECRET", originalS3Secret)
		} else {
			os.Unsetenv("STORAGE_S3_SECRET")
		}
	}()

	os.Setenv("STORAGE_TYPE", "s3")
	os.Setenv("STORAGE_S3_BUCKET", "test-bucket")
	os.Setenv("STORAGE_S3_REGION", "us-east-1")
	os.Setenv("STORAGE_S3_KEY", "test-key")
	os.Setenv("STORAGE_S3_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Storage.Type != "s3" {
		t.Errorf("Expected storage type 's3', got %q", cfg.Storage.Type)
	}
	if cfg.Storage.S3Bucket != "test-bucket" {
		t.Errorf("Expected S3 bucket 'test-bucket', got %q", cfg.Storage.S3Bucket)
	}
}

func TestValidate_JWTSecretKey_TooShort(t *testing.T) {
	// Save original values
	originalEnv := os.Getenv("ENV")
	originalJWT := os.Getenv("JWT_SECRET_KEY")
	originalCORS := os.Getenv("CORS_ALLOWED_ORIGINS")

	defer func() {
		if originalEnv != "" {
			os.Setenv("ENV", originalEnv)
		} else {
			os.Unsetenv("ENV")
		}
		if originalJWT != "" {
			os.Setenv("JWT_SECRET_KEY", originalJWT)
		} else {
			os.Unsetenv("JWT_SECRET_KEY")
		}
		if originalCORS != "" {
			os.Setenv("CORS_ALLOWED_ORIGINS", originalCORS)
		} else {
			os.Unsetenv("CORS_ALLOWED_ORIGINS")
		}
	}()

	os.Setenv("ENV", "production")
	os.Setenv("JWT_SECRET_KEY", "short") // Too short
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for JWT_SECRET_KEY too short, got nil")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T: %v", err, err)
	}

	errorMsg := validationErr.Error()
	if !strings.Contains(errorMsg, "32 characters") {
		t.Errorf("Expected validation error about JWT_SECRET_KEY length, got: %v", validationErr)
	}
}

func TestGetRequiredEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    []string
	}{
		{
			name:        "Development - no required vars",
			environment: "development",
			expected:    []string{},
		},
		{
			name:        "Staging - JWT_SECRET_KEY required",
			environment: "staging",
			expected:    []string{"JWT_SECRET_KEY"},
		},
		{
			name:        "Production - JWT_SECRET_KEY and CORS_ALLOWED_ORIGINS required",
			environment: "production",
			expected:    []string{"JWT_SECRET_KEY", "CORS_ALLOWED_ORIGINS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRequiredEnvVars(tt.environment)
			if len(result) != len(tt.expected) {
				t.Errorf("GetRequiredEnvVars(%q) length = %d, want %d", tt.environment, len(result), len(tt.expected))
				return
			}

			for _, expectedVar := range tt.expected {
				if !contains(result, expectedVar) {
					t.Errorf("GetRequiredEnvVars(%q) missing %q, got %v", tt.environment, expectedVar, result)
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Message: "validation failed",
		Fields:  []string{"field1", "field2"},
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "validation failed") {
		t.Errorf("Expected error message to contain 'validation failed', got: %q", errorMsg)
	}
	if !strings.Contains(errorMsg, "field1") || !strings.Contains(errorMsg, "field2") {
		t.Errorf("Expected error message to contain field names, got: %q", errorMsg)
	}
}

func TestRequiredEnvError_Error(t *testing.T) {
	err := &RequiredEnvError{
		Variables:  []string{"VAR1", "VAR2"},
		Environment: "production",
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "production") {
		t.Errorf("Expected error message to contain 'production', got: %q", errorMsg)
	}
	if !strings.Contains(errorMsg, "VAR1") || !strings.Contains(errorMsg, "VAR2") {
		t.Errorf("Expected error message to contain variable names, got: %q", errorMsg)
	}
}

func TestLoad_MetricsConfig_Defaults(t *testing.T) {
	// Save original values
	originalMetricsEnabled := os.Getenv("METRICS_ENABLED")
	originalMetricsPath := os.Getenv("METRICS_PATH")
	originalMetricsEnableHTTP := os.Getenv("METRICS_ENABLE_HTTP")
	originalMetricsEnableSystem := os.Getenv("METRICS_ENABLE_SYSTEM")
	originalMetricsEnableBusiness := os.Getenv("METRICS_ENABLE_BUSINESS")

	defer func() {
		if originalMetricsEnabled != "" {
			os.Setenv("METRICS_ENABLED", originalMetricsEnabled)
		} else {
			os.Unsetenv("METRICS_ENABLED")
		}
		if originalMetricsPath != "" {
			os.Setenv("METRICS_PATH", originalMetricsPath)
		} else {
			os.Unsetenv("METRICS_PATH")
		}
		if originalMetricsEnableHTTP != "" {
			os.Setenv("METRICS_ENABLE_HTTP", originalMetricsEnableHTTP)
		} else {
			os.Unsetenv("METRICS_ENABLE_HTTP")
		}
		if originalMetricsEnableSystem != "" {
			os.Setenv("METRICS_ENABLE_SYSTEM", originalMetricsEnableSystem)
		} else {
			os.Unsetenv("METRICS_ENABLE_SYSTEM")
		}
		if originalMetricsEnableBusiness != "" {
			os.Setenv("METRICS_ENABLE_BUSINESS", originalMetricsEnableBusiness)
		} else {
			os.Unsetenv("METRICS_ENABLE_BUSINESS")
		}
	}()

	// Unset all metrics environment variables to test defaults
	os.Unsetenv("METRICS_ENABLED")
	os.Unsetenv("METRICS_PATH")
	os.Unsetenv("METRICS_ENABLE_HTTP")
	os.Unsetenv("METRICS_ENABLE_SYSTEM")
	os.Unsetenv("METRICS_ENABLE_BUSINESS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check default values
	if cfg.Metrics.Enabled != true {
		t.Errorf("Expected Metrics.Enabled = true (default), got %v", cfg.Metrics.Enabled)
	}
	if cfg.Metrics.Path != "/metrics" {
		t.Errorf("Expected Metrics.Path = \"/metrics\" (default), got %q", cfg.Metrics.Path)
	}
	if cfg.Metrics.EnableHTTPMetrics != true {
		t.Errorf("Expected Metrics.EnableHTTPMetrics = true (default), got %v", cfg.Metrics.EnableHTTPMetrics)
	}
	if cfg.Metrics.EnableSystemMetrics != true {
		t.Errorf("Expected Metrics.EnableSystemMetrics = true (default), got %v", cfg.Metrics.EnableSystemMetrics)
	}
	if cfg.Metrics.EnableBusinessMetrics != true {
		t.Errorf("Expected Metrics.EnableBusinessMetrics = true (default), got %v", cfg.Metrics.EnableBusinessMetrics)
	}
}

func TestLoad_MetricsConfig_CustomValues(t *testing.T) {
	// Save original values
	originalMetricsEnabled := os.Getenv("METRICS_ENABLED")
	originalMetricsPath := os.Getenv("METRICS_PATH")
	originalMetricsEnableHTTP := os.Getenv("METRICS_ENABLE_HTTP")
	originalMetricsEnableSystem := os.Getenv("METRICS_ENABLE_SYSTEM")
	originalMetricsEnableBusiness := os.Getenv("METRICS_ENABLE_BUSINESS")

	defer func() {
		if originalMetricsEnabled != "" {
			os.Setenv("METRICS_ENABLED", originalMetricsEnabled)
		} else {
			os.Unsetenv("METRICS_ENABLED")
		}
		if originalMetricsPath != "" {
			os.Setenv("METRICS_PATH", originalMetricsPath)
		} else {
			os.Unsetenv("METRICS_PATH")
		}
		if originalMetricsEnableHTTP != "" {
			os.Setenv("METRICS_ENABLE_HTTP", originalMetricsEnableHTTP)
		} else {
			os.Unsetenv("METRICS_ENABLE_HTTP")
		}
		if originalMetricsEnableSystem != "" {
			os.Setenv("METRICS_ENABLE_SYSTEM", originalMetricsEnableSystem)
		} else {
			os.Unsetenv("METRICS_ENABLE_SYSTEM")
		}
		if originalMetricsEnableBusiness != "" {
			os.Setenv("METRICS_ENABLE_BUSINESS", originalMetricsEnableBusiness)
		} else {
			os.Unsetenv("METRICS_ENABLE_BUSINESS")
		}
	}()

	// Set custom values
	os.Setenv("METRICS_ENABLED", "false")
	os.Setenv("METRICS_PATH", "/custom-metrics")
	os.Setenv("METRICS_ENABLE_HTTP", "false")
	os.Setenv("METRICS_ENABLE_SYSTEM", "false")
	os.Setenv("METRICS_ENABLE_BUSINESS", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check custom values
	if cfg.Metrics.Enabled != false {
		t.Errorf("Expected Metrics.Enabled = false, got %v", cfg.Metrics.Enabled)
	}
	if cfg.Metrics.Path != "/custom-metrics" {
		t.Errorf("Expected Metrics.Path = \"/custom-metrics\", got %q", cfg.Metrics.Path)
	}
	if cfg.Metrics.EnableHTTPMetrics != false {
		t.Errorf("Expected Metrics.EnableHTTPMetrics = false, got %v", cfg.Metrics.EnableHTTPMetrics)
	}
	if cfg.Metrics.EnableSystemMetrics != false {
		t.Errorf("Expected Metrics.EnableSystemMetrics = false, got %v", cfg.Metrics.EnableSystemMetrics)
	}
	if cfg.Metrics.EnableBusinessMetrics != false {
		t.Errorf("Expected Metrics.EnableBusinessMetrics = false, got %v", cfg.Metrics.EnableBusinessMetrics)
	}
}

// Helper function to check if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
