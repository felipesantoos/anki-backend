package middlewares

import (
	"context"

	"github.com/labstack/echo/v4"
)

// RequestIDMiddleware returns an Echo middleware that generates or extracts a request ID
// and stores it in the context for use with GetRequestID().
// It checks for an existing X-Request-ID header first, and generates a new one if not present.
// The Request ID is also added to the response headers.
//
// This middleware should be registered early in the middleware chain (ideally as the first middleware)
// to ensure the Request ID is available to all subsequent middlewares and handlers.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract or generate Request ID
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add request ID to context using requestIDKey (for compatibility with GetRequestID())
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, requestIDKey{}, requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Add request ID to response header
			c.Response().Header().Set("X-Request-ID", requestID)

			return next(c)
		}
	}
}
