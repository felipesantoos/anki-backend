package routes

import (
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/dicontainer"
)

// RegisterMaintenanceRoutes registers maintenance-related routes on the Router
func (r *Router) RegisterMaintenanceRoutes() {
	cardService := dicontainer.GetCardService()
	handler := handlers.NewMaintenanceHandler(cardService)

	// Create maintenance group
	maintenanceGroup := r.echo.Group("/api/v1/maintenance")
	
	// Create authenticated maintenance group (requires JWT)
	authMiddleware := middlewares.AuthMiddleware(r.jwtSvc, r.rdb)
	maintenanceGroup.Use(authMiddleware)

	// Register routes
	maintenanceGroup.GET("/empty-cards", handler.GetEmptyCards)
	maintenanceGroup.POST("/empty-cards/cleanup", handler.CleanupEmptyCards)
}
