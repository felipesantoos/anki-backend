package routes

import (
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/dicontainer"
)

// RegisterHealthRoutes registers health check routes on the Router
func (r *Router) RegisterHealthRoutes() {
	healthService := dicontainer.GetHealthService()
	healthHandler := handlers.NewHealthHandler(healthService)

	// Health check endpoint
	r.echo.GET("/health", healthHandler.HealthCheck)
}
