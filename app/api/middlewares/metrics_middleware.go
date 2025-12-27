package middlewares

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// normalizePath normalizes HTTP paths to avoid high cardinality in metrics
// Examples:
//   - "/api/users/123" -> "/api/users/:id"
//   - "/api/decks/456/cards/789" -> "/api/decks/:id/cards/:id"
//   - "/health" -> "/health"
func normalizePath(path string) string {
	// Replace UUIDs (8-4-4-4-12 format)
	path = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).ReplaceAllString(path, ":id")
	
	// Replace numeric IDs (sequences of digits)
	path = regexp.MustCompile(`/\d+`).ReplaceAllString(path, "/:id")
	
	// Replace query strings
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	
	return path
}

// MetricsMiddleware creates a middleware that collects HTTP metrics
// It should be placed after RequestID middleware to have request IDs in logs
func MetricsMiddleware(metricsService primary.IMetricsService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip metrics collection if service is nil
			if metricsService == nil {
				return next(c)
			}

			start := time.Now()
			req := c.Request()
			method := req.Method
			path := normalizePath(req.URL.Path)

			// Get request size
			var requestSize int64
			if req.ContentLength > 0 {
				requestSize = req.ContentLength
			}

			// Process the request
			err := next(c)

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Get status code
			statusCode := c.Response().Status
			if statusCode == 0 {
				statusCode = 200 // Default status if not set
			}
			statusStr := strconv.Itoa(statusCode)

			// Get response size
			responseSize := int64(c.Response().Size)

			// Record metrics using the service method
			metricsService.RecordHTTPRequest(method, path, statusStr, duration, requestSize, responseSize)

			return err
		}
	}
}



