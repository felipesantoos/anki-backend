package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// memoryFixedWindowLimiter implements fixed window rate limiting in memory
// Similar to Redis fixed window, but stores counters in a map with automatic cleanup
type memoryFixedWindowLimiter struct {
	counters map[string]*windowCounter
	mu       sync.RWMutex
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

type windowCounter struct {
	count     int
	windowEnd int64 // Unix timestamp when window expires
}

func newMemoryFixedWindowLimiter() *memoryFixedWindowLimiter {
	limiter := &memoryFixedWindowLimiter{
		counters: make(map[string]*windowCounter),
		cleanupInterval: 5 * time.Minute, // Clean up old entries every 5 minutes
		lastCleanup:     time.Now(),
	}
	return limiter
}

// allow checks if a request should be allowed using fixed window algorithm
// Returns (allowed, remaining, resetTime)
func (m *memoryFixedWindowLimiter) allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time) {
	now := time.Now()
	windowStart := now.Truncate(window)
	windowEnd := windowStart.Add(window)
	windowEndUnix := windowEnd.Unix()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Cleanup old entries periodically to prevent memory leak (while holding lock)
	m.cleanupIfNeededUnsafe(now)

	// Get or create counter for this key and window
	counter, exists := m.counters[key]
	if !exists || counter.windowEnd != windowEndUnix {
		// New window or new key - start fresh
		counter = &windowCounter{
			count:     0,
			windowEnd: windowEndUnix,
		}
		m.counters[key] = counter
	}

	// Increment counter
	counter.count++

	remaining := limit - counter.count
	if remaining < 0 {
		remaining = 0
	}

	allowed := counter.count <= limit
	resetTime := windowEnd

	return allowed, remaining, resetTime
}

// cleanupIfNeededUnsafe removes expired entries to prevent memory leaks
// MUST be called while holding m.mu lock
func (m *memoryFixedWindowLimiter) cleanupIfNeededUnsafe(now time.Time) {
	if now.Sub(m.lastCleanup) < m.cleanupInterval {
		return
	}

	// Remove expired entries (windows that ended more than cleanupInterval ago)
	cutoff := now.Add(-m.cleanupInterval).Unix()
	for key, counter := range m.counters {
		if counter.windowEnd < cutoff {
			delete(m.counters, key)
		}
	}

	m.lastCleanup = now
}

// redisLimiter uses Redis for distributed rate limiting
type redisLimiter struct {
	client *redis.Client
}

func newRedisLimiter(client *redis.Client) *redisLimiter {
	return &redisLimiter{
		client: client,
	}
}

// allow checks if a request should be allowed using Redis fixed window algorithm
// Returns (allowed, remaining, resetTime)
func (r *redisLimiter) allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Truncate(window).Unix())
	
	// Increment counter in Redis
	count, err := r.client.Incr(ctx, windowKey).Result()
	if err != nil {
		// If Redis fails, allow the request but log the error
		logger.GetLogger().Warn("Redis rate limit check failed, allowing request", "error", err, "key", key)
		return true, limit, now.Add(window)
	}

	// Set expiration on first request
	if count == 1 {
		r.client.Expire(ctx, windowKey, window)
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	allowed := count <= int64(limit)
	resetTime := now.Truncate(window).Add(window)

	return allowed, remaining, resetTime
}

// rateLimitStore is an interface for rate limiting storage backends
type rateLimitStore interface {
	allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time)
}

// memoryRateLimitStore adapts memoryFixedWindowLimiter to rateLimitStore interface
type memoryRateLimitStore struct {
	limiter *memoryFixedWindowLimiter
	defaultLimit int
	window  time.Duration
}

func (m *memoryRateLimitStore) allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time) {
	// Use the provided limit, or default if not specified
	effectiveLimit := limit
	if effectiveLimit <= 0 {
		effectiveLimit = m.defaultLimit
	}
	
	// Use fixed window algorithm (same as Redis)
	return m.limiter.allow(ctx, key, effectiveLimit, window)
}

// redisRateLimitStore adapts redisLimiter to rateLimitStore interface
type redisRateLimitStore struct {
	limiter *redisLimiter
}

func (r *redisRateLimitStore) allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time) {
	return r.limiter.allow(ctx, key, limit, window)
}

// getEndpointLimit returns the rate limit for a specific endpoint
func getEndpointLimit(cfg config.RateLimitConfig, path string) int {
	if !cfg.EnableByEndpoint {
		return cfg.DefaultLimitPerMinute
	}

	// Check for specific endpoints
	pathLower := strings.ToLower(path)
	if strings.Contains(pathLower, "login") || strings.Contains(pathLower, "auth/login") {
		return cfg.LoginLimitPerMinute
	}

	return cfg.DefaultLimitPerMinute
}

// isExcludedPath checks if a path should be excluded from rate limiting
func isExcludedPath(path string) bool {
	excludedPaths := []string{
		"/health",
		"/swagger",
	}

	for _, excluded := range excludedPaths {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}

	return false
}

// RateLimitingMiddleware creates a rate limiting middleware for Echo
// It supports both Redis (distributed) and memory (single instance) strategies
func RateLimitingMiddleware(cfg config.RateLimitConfig, redisClient *redis.Client) echo.MiddlewareFunc {
	// If rate limiting is disabled, return a no-op middleware
	if !cfg.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	var store rateLimitStore

	// Initialize store based on strategy
	if cfg.Strategy == "redis" && redisClient != nil {
		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := redisClient.Ping(ctx).Err()
		cancel()

		if err != nil {
			// Redis unavailable, fallback to memory
			logger.GetLogger().Warn("Redis unavailable for rate limiting, falling back to memory strategy", "error", err)
			memLimiter := newMemoryFixedWindowLimiter()
			store = &memoryRateLimitStore{
				limiter:      memLimiter,
				defaultLimit: cfg.DefaultLimitPerMinute,
				window:       time.Minute,
			}
		} else {
			// Redis available, use it
			redisLimiter := newRedisLimiter(redisClient)
			store = &redisRateLimitStore{limiter: redisLimiter}
		}
	} else {
		// Use memory strategy (fixed window, same algorithm as Redis)
		memLimiter := newMemoryFixedWindowLimiter()
		store = &memoryRateLimitStore{
			limiter:      memLimiter,
			defaultLimit: cfg.DefaultLimitPerMinute,
			window:       time.Minute,
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Skip rate limiting for excluded paths
			if isExcludedPath(path) {
				return next(c)
			}

			// Get client identifier (IP address)
			clientIP := getClientIP(c.Request())
			if clientIP == "" {
				clientIP = "unknown"
			}

			// Get endpoint-specific limit
			limit := getEndpointLimit(cfg, path)

			// Create key: use path for endpoint-specific limits, or "global" for default
			key := clientIP
			if cfg.EnableByEndpoint {
				key = fmt.Sprintf("%s:%s", clientIP, path)
			}

			// Check rate limit (pass the endpoint-specific limit)
			allowed, remaining, resetTime := store.allow(c.Request().Context(), key, limit, time.Minute)

			// Set rate limit headers (RFC 6585)
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				log := logger.GetLogger()
				log.Warn("Rate limit exceeded",
					"client_ip", clientIP,
					"path", path,
					"limit", limit,
					"request_id", GetRequestID(c.Request().Context()),
				)

				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":   "Too Many Requests",
					"message": "Rate limit exceeded. Please try again later.",
					"retry_after": int(time.Until(resetTime).Seconds()),
				})
			}

			return next(c)
		}
	}
}
