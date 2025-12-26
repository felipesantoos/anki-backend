package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// RegisterHealthRoutes registers health check routes
func RegisterHealthRoutes(e *echo.Echo, healthService primary.IHealthService) {
	healthHandler := handlers.NewHealthHandler(healthService)

	// Health check endpoint
	e.GET("/health", healthHandler.HealthCheck)
}
