package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/felipesantos/anki-backend/config"
)

// CORSMiddleware returns an Echo middleware that configures CORS (Cross-Origin Resource Sharing)
// based on the provided configuration. It uses Echo's built-in CORS middleware with
// configurable allowed origins, methods, headers, and credentials.
//
// If CORS is disabled in the configuration, this middleware returns a no-op middleware.
//
// Configuration:
//   - AllowOrigins: List of allowed origins (from config.CORS.AllowedOrigins)
//   - AllowMethods: Standard HTTP methods [GET, POST, PUT, DELETE, PATCH, OPTIONS]
//   - AllowHeaders: Common headers [Content-Type, Authorization, X-Request-ID]
//   - AllowCredentials: Whether to allow credentials (from config.CORS.AllowCredentials)
//   - MaxAge: 12 hours (43200 seconds) - how long browsers cache preflight responses
//
// Security notes:
//   - In production, never use "*" as origin if AllowCredentials=true
//   - Always specify explicit origins in production environments
func CORSMiddleware(cfg config.CORSConfig) echo.MiddlewareFunc {
	// If CORS is disabled, return a no-op middleware
	if !cfg.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	// Configure CORS middleware with standard settings
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.DELETE,
			echo.PATCH,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderContentType,
			echo.HeaderAuthorization,
			"X-Request-ID",
		},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           43200, // 12 hours in seconds
	})
}
