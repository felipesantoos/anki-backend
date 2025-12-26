package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/felipesantos/anki-backend/pkg/logger"
)

// RequestIDKey is the key used to store the request ID in the context
type requestIDKey struct{}

// GetRequestID extracts the request ID from the context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// responseWriter wrapper to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// generateRequestID generates a unique ID for the request
func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// getClientIP extracts the client IP from the request
// Considers proxy/load balancer headers (X-Forwarded-For, X-Real-IP)
// Returns only the IP address, removing the port number
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For (proxies/load balancers)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs (comma-separated), take the first one
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = strings.TrimSpace(ip[:idx])
		}
		// Remove port if present
		return normalizeIP(ip)
	}
	// Check X-Real-IP
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return normalizeIP(ip)
	}
	// Fallback to RemoteAddr (remove port)
	return normalizeIP(r.RemoteAddr)
}

// normalizeIP removes the port number from an IP address
// Examples: "[::1]:12345" -> "::1", "127.0.0.1:8080" -> "127.0.0.1", "::1:12345" -> "::1"
func normalizeIP(addr string) string {
	// Use net.SplitHostPort which properly handles both IPv4 and IPv6
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// If SplitHostPort fails, assume it's already just an IP without port
		return addr
	}
	return host
}

// LoggingMiddleware logs all HTTP requests and responses
// Returns a middleware function that can be used with http.Handler
func LoggingMiddleware() func(http.Handler) http.Handler {
	log := logger.GetLogger()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract or generate Request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add request ID to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, requestIDKey{}, requestID)
			r = r.WithContext(ctx)

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Create response writer wrapper
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status code
			}

			// Log request start
			log.Info("Request started",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"client_ip", getClientIP(r),
				"user_agent", r.UserAgent(),
			)

			// Process request with panic handling
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						// Calculate duration until panic
						duration := time.Since(start)

						// Log panic
						log.Error("Request panicked",
							"request_id", requestID,
							"method", r.Method,
							"path", r.URL.Path,
							"panic", rec,
							"duration_ms", duration.Milliseconds(),
						)

						// Return 500 error if not yet written
						if rw.statusCode == http.StatusOK {
							rw.WriteHeader(http.StatusInternalServerError)
						}
					}
				}()

				next.ServeHTTP(rw, r)
			}()

			// Calculate duration
			duration := time.Since(start)

			// Log request end
			log.Info("Request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes_written", rw.bytesWritten,
			)
		})
	}
}

// LoggingMiddlewareWithLogger allows passing a custom logger
// Useful for tests or when you want to use a specific logger
func LoggingMiddlewareWithLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Extract or generate Request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add request ID to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, requestIDKey{}, requestID)
			r = r.WithContext(ctx)

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Create response writer wrapper
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Log request start
			log.Info("Request started",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"client_ip", getClientIP(r),
				"user_agent", r.UserAgent(),
			)

			// Process request with panic handling
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						// Calculate duration until panic
						duration := time.Since(start)

						// Log panic
						log.Error("Request panicked",
							"request_id", requestID,
							"method", r.Method,
							"path", r.URL.Path,
							"panic", rec,
							"duration_ms", duration.Milliseconds(),
						)

						// Return 500 error if not yet written
						if rw.statusCode == http.StatusOK {
							rw.WriteHeader(http.StatusInternalServerError)
						}
					}
				}()

				next.ServeHTTP(rw, r)
			}()

			// Calculate duration
			duration := time.Since(start)

			// Log request end
			log.Info("Request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes_written", rw.bytesWritten,
			)
		})
	}
}

