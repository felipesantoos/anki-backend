package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// Redis configuration
	Redis RedisConfig

	// JWT configuration
	JWT JWTConfig

	// Storage configuration
	Storage StorageConfig

	// Logger configuration
	Logger LoggerConfig

	// Rate limiting configuration
	RateLimit RateLimitConfig

	// CORS configuration
	CORS CORSConfig

	// Session configuration
	Session SessionConfig

	// Jobs configuration
	Jobs JobsConfig

	// Events configuration
	Events EventsConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     int
	WriteTimeout    int
	IdleTimeout     int
	ShutdownTimeout int // Graceful shutdown timeout in seconds (default: 10)
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host             string
	Port             string
	User             string
	Password         string
	DBName           string
	SSLMode          string
	MaxConnections   int
	MaxIdleConns     int
	ConnMaxLifetime  int
	ConnMaxIdleTime  int // Maximum time a connection can be idle before being closed (in minutes)
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SecretKey           string
	AccessTokenExpiry   int
	RefreshTokenExpiry  int
	Issuer              string
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	Type              string // "local", "s3", "cloudflare", or "r2"
	LocalPath         string
	S3Bucket          string
	S3Region          string
	S3Key             string
	S3Secret          string
	CloudflareAccountID string // Cloudflare Account ID for R2
	CloudflareR2Bucket  string // Cloudflare R2 bucket name
	CloudflareR2Key     string // Cloudflare R2 Access Key ID
	CloudflareR2Secret  string // Cloudflare R2 Secret Access Key
	CloudflareR2Endpoint string // Cloudflare R2 endpoint (optional, defaults to https://<account-id>.r2.cloudflarestorage.com)
}

// LoggerConfig holds logger-related configuration
type LoggerConfig struct {
	Level       string // "debug", "info", "warn", "error"
	Environment string // "development", "staging", "production"
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled              bool   // Enable/disable rate limiting
	Strategy             string // "redis" or "memory"
	DefaultLimitPerMinute int   // Default limit per minute (e.g., 60)
	Burst                int    // Burst allowed (e.g., 10 extra requests)
	EnableByEndpoint     bool   // Enable per-endpoint limits
	LoginLimitPerMinute  int    // Login endpoint limit per minute (e.g., 5)
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled         bool     // Enable/disable CORS middleware
	AllowedOrigins  []string // List of allowed origins (comma-separated in env var)
	AllowCredentials bool    // Allow credentials (cookies, authorization headers)
}

// SessionConfig holds session-related configuration
type SessionConfig struct {
	TTLMinutes int    // Session time-to-live in minutes (default: 30)
	KeyPrefix  string // Prefix for session keys in Redis (default: "session")
}

// JobsConfig holds background jobs configuration
type JobsConfig struct {
	Enabled          bool   // Enable/disable job processing system
	WorkerCount      int    // Number of worker goroutines (default: 5)
	QueueSize        int    // Queue buffer size (default: 1000)
	MaxRetries       int    // Maximum number of retries for failed jobs (default: 3)
	RetryDelaySeconds int   // Base delay between retries in seconds (default: 5)
	RedisQueueKey    string // Redis key for job queue (default: "jobs:queue")
	RedisDB          int    // Redis database number for jobs (default: 1, use 0 for same as cache)
}

// EventsConfig holds domain events bus configuration
type EventsConfig struct {
	Enabled    bool // Enable/disable event bus
	WorkerCount int // Number of workers to process events
	QueueSize   int // Size of the event queue buffer
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Message string
	Fields  []string
}

func (e *ValidationError) Error() string {
	if len(e.Fields) > 0 {
		return fmt.Sprintf("%s: %s", e.Message, strings.Join(e.Fields, ", "))
	}
	return e.Message
}

// RequiredEnvError represents an error for missing required environment variables
type RequiredEnvError struct {
	Variables []string
	Environment string
}

func (e *RequiredEnvError) Error() string {
	varsList := strings.Join(e.Variables, ", ")
	return fmt.Sprintf("missing required environment variables for %s environment: %s", e.Environment, varsList)
}

// Load loads configuration from environment variables
// It automatically tries to load .env file if it exists (does not fail if missing)
// Environment variables take precedence over .env file values
func Load() (*Config, error) {
	// Try to load .env file (silently ignore if it doesn't exist)
	_ = godotenv.Load()

	return loadConfig()
}

// LoadFromFile loads configuration from a specific .env file
// Environment variables take precedence over file values
func LoadFromFile(filename string) error {
	return godotenv.Load(filename)
}

// loadConfig performs the actual configuration loading
func loadConfig() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout:    getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:     getEnvAsInt("SERVER_IDLE_TIMEOUT", 120),
			ShutdownTimeout: getEnvAsInt("SERVER_SHUTDOWN_TIMEOUT", 10),
		},
		Database: DatabaseConfig{
			Host:             getEnv("DB_HOST", "localhost"),
			Port:             getEnv("DB_PORT", "5432"),
			User:             getEnv("DB_USER", "postgres"),
			Password:         getEnv("DB_PASSWORD", ""),
			DBName:           getEnv("DB_NAME", "anki"),
			SSLMode:          getEnv("DB_SSLMODE", "disable"),
			MaxConnections:   getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:     getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime:  getEnvAsInt("DB_CONN_MAX_LIFETIME", 5),
			ConnMaxIdleTime:  getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 10),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			SecretKey:          getEnv("JWT_SECRET_KEY", ""),
			AccessTokenExpiry:  getEnvAsInt("JWT_ACCESS_TOKEN_EXPIRY", 15),
			RefreshTokenExpiry: getEnvAsInt("JWT_REFRESH_TOKEN_EXPIRY", 7),
			Issuer:             getEnv("JWT_ISSUER", "anki-api"),
		},
		Storage: StorageConfig{
			Type:                getEnv("STORAGE_TYPE", "local"),
			LocalPath:           getEnv("STORAGE_LOCAL_PATH", "./storage"),
			S3Bucket:            getEnv("STORAGE_S3_BUCKET", ""),
			S3Region:            getEnv("STORAGE_S3_REGION", ""),
			S3Key:               getEnv("STORAGE_S3_KEY", ""),
			S3Secret:            getEnv("STORAGE_S3_SECRET", ""),
			CloudflareAccountID: getEnv("STORAGE_CLOUDFLARE_ACCOUNT_ID", ""),
			CloudflareR2Bucket:  getEnv("STORAGE_CLOUDFLARE_R2_BUCKET", ""),
			CloudflareR2Key:     getEnv("STORAGE_CLOUDFLARE_R2_KEY", ""),
			CloudflareR2Secret:  getEnv("STORAGE_CLOUDFLARE_R2_SECRET", ""),
			CloudflareR2Endpoint: getEnv("STORAGE_CLOUDFLARE_R2_ENDPOINT", ""),
		},
		Logger: LoggerConfig{
			Level:       validateLogLevel(getEnv("LOG_LEVEL", "info")),
			Environment: validateEnvironment(getEnv("ENV", "development")),
		},
		RateLimit: RateLimitConfig{
			Enabled:              getEnvAsBool("RATE_LIMIT_ENABLED", true),
			Strategy:             validateRateLimitStrategy(getEnv("RATE_LIMIT_STRATEGY", "redis")),
			DefaultLimitPerMinute: getEnvAsInt("RATE_LIMIT_DEFAULT_PER_MINUTE", 60),
			Burst:                getEnvAsInt("RATE_LIMIT_BURST", 10),
			EnableByEndpoint:     getEnvAsBool("RATE_LIMIT_ENABLE_BY_ENDPOINT", false),
			LoginLimitPerMinute:  getEnvAsInt("RATE_LIMIT_LOGIN_PER_MINUTE", 5),
		},
	}

	// Load CORS configuration
	// Default allowed origins is "*" for development, should be configured explicitly in production
	env := validateEnvironment(getEnv("ENV", "development"))
	corsDefaultOrigins := "*"
	if env == "production" {
		// In production, require explicit configuration (empty by default)
		corsDefaultOrigins = ""
	}

	cfg.CORS = CORSConfig{
		Enabled:          getEnvAsBool("CORS_ENABLED", true),
		AllowedOrigins:   parseCORSOrigins(getEnv("CORS_ALLOWED_ORIGINS", corsDefaultOrigins)),
		AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
	}

	cfg.Session = SessionConfig{
		TTLMinutes: getEnvAsInt("SESSION_TTL_MINUTES", 30),
		KeyPrefix:  getEnv("SESSION_KEY_PREFIX", "session"),
	}

	cfg.Jobs = JobsConfig{
		Enabled:           getEnvAsBool("JOBS_ENABLED", true),
		WorkerCount:       getEnvAsInt("JOBS_WORKER_COUNT", 5),
		QueueSize:         getEnvAsInt("JOBS_QUEUE_SIZE", 1000),
		MaxRetries:        getEnvAsInt("JOBS_MAX_RETRIES", 3),
		RetryDelaySeconds: getEnvAsInt("JOBS_RETRY_DELAY_SECONDS", 5),
		RedisQueueKey:     getEnv("JOBS_REDIS_QUEUE_KEY", "jobs:queue"),
		RedisDB:           getEnvAsInt("JOBS_REDIS_DB", 1),
	}

	cfg.Events = EventsConfig{
		Enabled:    getEnvAsBool("EVENTS_ENABLED", true),
		WorkerCount: getEnvAsInt("EVENTS_WORKER_COUNT", 5),
		QueueSize:   getEnvAsInt("EVENTS_QUEUE_SIZE", 1000),
	}

	// Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration and returns an error if validation fails
func Validate(cfg *Config) error {
	var missingVars []string
	var validationErrors []string

	env := cfg.Logger.Environment

	// Validate JWT Secret Key (required in staging/production)
	if env == "staging" || env == "production" {
		if cfg.JWT.SecretKey == "" {
			missingVars = append(missingVars, "JWT_SECRET_KEY")
			validationErrors = append(validationErrors, "JWT_SECRET_KEY is required in staging/production environments")
		} else if len(cfg.JWT.SecretKey) < 32 {
			validationErrors = append(validationErrors, "JWT_SECRET_KEY must be at least 32 characters long for security")
		}
	}

	// Validate CORS configuration in production
	if env == "production" {
		if len(cfg.CORS.AllowedOrigins) == 0 {
			missingVars = append(missingVars, "CORS_ALLOWED_ORIGINS")
			validationErrors = append(validationErrors, "CORS_ALLOWED_ORIGINS is required in production and cannot be empty")
		} else {
			// Check if wildcard is used (not allowed in production with credentials)
			for _, origin := range cfg.CORS.AllowedOrigins {
				if origin == "*" && cfg.CORS.AllowCredentials {
					validationErrors = append(validationErrors, "CORS wildcard (*) is not allowed in production when AllowCredentials is true")
					break
				}
			}
		}
	}

	// Validate S3 configuration if storage type is s3
	if cfg.Storage.Type == "s3" {
		if cfg.Storage.S3Bucket == "" {
			missingVars = append(missingVars, "STORAGE_S3_BUCKET")
			validationErrors = append(validationErrors, "STORAGE_S3_BUCKET is required when STORAGE_TYPE=s3")
		}
		if cfg.Storage.S3Region == "" {
			missingVars = append(missingVars, "STORAGE_S3_REGION")
			validationErrors = append(validationErrors, "STORAGE_S3_REGION is required when STORAGE_TYPE=s3")
		}
		if cfg.Storage.S3Key == "" {
			missingVars = append(missingVars, "STORAGE_S3_KEY")
			validationErrors = append(validationErrors, "STORAGE_S3_KEY is required when STORAGE_TYPE=s3")
		}
		if cfg.Storage.S3Secret == "" {
			missingVars = append(missingVars, "STORAGE_S3_SECRET")
			validationErrors = append(validationErrors, "STORAGE_S3_SECRET is required when STORAGE_TYPE=s3")
		}
	}

	// Validate Cloudflare R2 configuration if storage type is cloudflare or r2
	if cfg.Storage.Type == "cloudflare" || cfg.Storage.Type == "r2" {
		if cfg.Storage.CloudflareAccountID == "" {
			missingVars = append(missingVars, "STORAGE_CLOUDFLARE_ACCOUNT_ID")
			validationErrors = append(validationErrors, "STORAGE_CLOUDFLARE_ACCOUNT_ID is required when STORAGE_TYPE=cloudflare or r2")
		}
		if cfg.Storage.CloudflareR2Bucket == "" {
			missingVars = append(missingVars, "STORAGE_CLOUDFLARE_R2_BUCKET")
			validationErrors = append(validationErrors, "STORAGE_CLOUDFLARE_R2_BUCKET is required when STORAGE_TYPE=cloudflare or r2")
		}
		if cfg.Storage.CloudflareR2Key == "" {
			missingVars = append(missingVars, "STORAGE_CLOUDFLARE_R2_KEY")
			validationErrors = append(validationErrors, "STORAGE_CLOUDFLARE_R2_KEY is required when STORAGE_TYPE=cloudflare or r2")
		}
		if cfg.Storage.CloudflareR2Secret == "" {
			missingVars = append(missingVars, "STORAGE_CLOUDFLARE_R2_SECRET")
			validationErrors = append(validationErrors, "STORAGE_CLOUDFLARE_R2_SECRET is required when STORAGE_TYPE=cloudflare or r2")
		}
	}

	// Return appropriate error type
	if len(missingVars) > 0 {
		return &RequiredEnvError{
			Variables:  missingVars,
			Environment: env,
		}
	}

	if len(validationErrors) > 0 {
		return &ValidationError{
			Message: "configuration validation failed",
			Fields:  validationErrors,
		}
	}

	return nil
}

// GetRequiredEnvVars returns a list of required environment variables for the given environment
func GetRequiredEnvVars(environment string) []string {
	env := validateEnvironment(environment)
	
	var required []string

	if env == "staging" || env == "production" {
		required = append(required, "JWT_SECRET_KEY")
	}

	if env == "production" {
		required = append(required, "CORS_ALLOWED_ORIGINS")
	}

	// Note: S3 variables are conditionally required based on STORAGE_TYPE
	// so they are not included in this list

	return required
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsBool gets an environment variable as boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// validateLogLevel validates and normalizes the log level
// Returns "info" if the value is invalid
func validateLogLevel(level string) string {
	validLevels := map[string]string{
		"debug":   "debug",
		"DEBUG":   "debug",
		"info":    "info",
		"INFO":    "info",
		"warn":    "warn",
		"WARN":    "warn",
		"warning": "warn",
		"WARNING": "warn",
		"error":   "error",
		"ERROR":   "error",
	}

	if normalized, ok := validLevels[level]; ok {
		return normalized
	}
	// Return safe default if invalid
	return "info"
}

// validateEnvironment validates and normalizes the environment
// Returns "development" if the value is invalid
func validateEnvironment(env string) string {
	validEnvs := map[string]string{
		"development": "development",
		"dev":         "development",
		"DEV":         "development",
		"staging":     "staging",
		"stage":       "staging",
		"STAGING":     "staging",
		"production":  "production",
		"prod":        "production",
		"PROD":        "production",
		"PRODUCTION":  "production",
	}

	if normalized, ok := validEnvs[env]; ok {
		return normalized
	}
	// Return safe default if invalid
	return "development"
}

// validateRateLimitStrategy validates and normalizes the rate limit strategy
// Returns "redis" if the value is invalid
func validateRateLimitStrategy(strategy string) string {
	validStrategies := map[string]string{
		"redis":  "redis",
		"REDIS":  "redis",
		"memory": "memory",
		"MEMORY": "memory",
	}

	if normalized, ok := validStrategies[strategy]; ok {
		return normalized
	}
	// Return safe default if invalid
	return "redis"
}

// parseCORSOrigins parses a comma-separated string of CORS origins
// Returns a slice of origins with trimmed whitespace
// If the string is "*", returns []string{"*"}
// If empty, returns []string{} (empty = no origins allowed)
func parseCORSOrigins(originsStr string) []string {
	originsStr = strings.TrimSpace(originsStr)

	// Handle wildcard
	if originsStr == "*" {
		return []string{"*"}
	}

	// Handle empty string
	if originsStr == "" {
		return []string{}
	}

	// Split by comma and trim each origin
	parts := strings.Split(originsStr, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return origins
}
