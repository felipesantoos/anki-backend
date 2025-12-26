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
	"golang.org/x/time/rate"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/pkg/logger"
)

// memoryLimiter stores rate limiters in memory (fallback when Redis is unavailable)
// It supports dynamic limits per key, allowing different endpoints to have different rate limits
type memoryLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	defaultLimit int
	defaultBurst int
}

func newMemoryLimiter(limitPerMinute int, burst int) *memoryLimiter {
	return &memoryLimiter{
		limiters: make(map[string]*rate.Limiter),
		defaultLimit: limitPerMinute,
		defaultBurst: burst,
	}
}

func (m *memoryLimiter) getLimiter(key string, limitPerMinute int, burst int) *rate.Limiter {
	// Use the provided limit and burst, or default if not specified
	effectiveLimit := limitPerMinute
	if effectiveLimit <= 0 {
		effectiveLimit = m.defaultLimit
	}
	effectiveBurst := burst
	if effectiveBurst <= 0 {
		effectiveBurst = m.defaultBurst
	}
	
	// For token bucket with rate limiting "X requests per minute", we want to allow
	// up to X requests immediately, then enforce the rate. This requires burst >= limit.
	// If burst is smaller than limit, set it to limit to match the expected behavior.
	if effectiveBurst < effectiveLimit {
		effectiveBurst = effectiveLimit
	}
	
	// Create a composite key that includes the limit to support different limits per endpoint
	limitKey := fmt.Sprintf("%s:limit:%d:burst:%d", key, effectiveLimit, effectiveBurst)
	
	m.mu.RLock()
	limiter, exists := m.limiters[limitKey]
	m.mu.RUnlock()

	if exists {
		return limiter
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := m.limiters[limitKey]; exists {
		return limiter
	}

	rateLimit := rate.Limit(float64(effectiveLimit) / 60.0) // Convert per minute to per second
	limiter = rate.NewLimiter(rateLimit, effectiveBurst)
	m.limiters[limitKey] = limiter
	return limiter
}

func (m *memoryLimiter) allow(key string, limitPerMinute int, burst int) bool {
	limiter := m.getLimiter(key, limitPerMinute, burst)
	return limiter.Allow()
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

// memoryRateLimitStore adapts memoryLimiter to rateLimitStore interface
type memoryRateLimitStore struct {
	limiter *memoryLimiter
	defaultLimit int
	defaultBurst int
	window  time.Duration
}

func (m *memoryRateLimitStore) allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time) {
	// Use the provided limit, or default if not specified
	effectiveLimit := limit
	if effectiveLimit <= 0 {
		effectiveLimit = m.defaultLimit
	}
	effectiveBurst := m.defaultBurst
	if effectiveBurst <= 0 {
		effectiveBurst = effectiveLimit // Use limit as minimum burst
	}
	
	// For token bucket with rate limiting "X requests per minute", we want to allow
	// up to X requests immediately, then enforce the rate. This requires burst >= limit.
	if effectiveBurst < effectiveLimit {
		effectiveBurst = effectiveLimit
	}
	
	allowed := m.limiter.allow(key, effectiveLimit, effectiveBurst)
	
	// For memory limiter, we can't easily calculate remaining without more complex state
	// We'll use a simplified approach: if allowed, assume we're within limit
	remaining := effectiveLimit
	if !allowed {
		remaining = 0
	}
	
	resetTime := time.Now().Add(window)
	return allowed, remaining, resetTime
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
			memLimiter := newMemoryLimiter(cfg.DefaultLimitPerMinute, cfg.Burst)
			store = &memoryRateLimitStore{
				limiter:      memLimiter,
				defaultLimit: cfg.DefaultLimitPerMinute,
				defaultBurst: cfg.Burst,
				window:       time.Minute,
			}
		} else {
			// Redis available, use it
			redisLimiter := newRedisLimiter(redisClient)
			store = &redisRateLimitStore{limiter: redisLimiter}
		}
	} else {
		// Use memory strategy
		memLimiter := newMemoryLimiter(cfg.DefaultLimitPerMinute, cfg.Burst)
		store = &memoryRateLimitStore{
			limiter:      memLimiter,
			defaultLimit: cfg.DefaultLimitPerMinute,
			defaultBurst: cfg.Burst,
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
