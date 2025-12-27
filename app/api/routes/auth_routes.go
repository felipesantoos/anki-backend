package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
)

// RegisterAuthRoutes registers authentication routes on the Echo router
func RegisterAuthRoutes(e *echo.Echo, authService primary.IAuthService) {
	authHandler := handlers.NewAuthHandler(authService)

	// Create auth group
	authGroup := e.Group("/api/v1/auth")

	// Register routes
	authGroup.POST("/register", authHandler.Register)
}
