package config

import (
	"os"
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

	// Test 1: Valid values
	os.Setenv("ENV", "production")
	os.Setenv("LOG_LEVEL", "WARN")

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

