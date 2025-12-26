package config

import (
	"os"
	"strconv"
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
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime int
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
	Type      string // "local" or "s3"
	LocalPath string
	S3Bucket  string
	S3Region  string
	S3Key     string
	S3Secret  string
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

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvAsInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "anki"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 5),
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
			Type:      getEnv("STORAGE_TYPE", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./storage"),
			S3Bucket:  getEnv("STORAGE_S3_BUCKET", ""),
			S3Region:  getEnv("STORAGE_S3_REGION", ""),
			S3Key:     getEnv("STORAGE_S3_KEY", ""),
			S3Secret:  getEnv("STORAGE_S3_SECRET", ""),
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

	return cfg, nil
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

