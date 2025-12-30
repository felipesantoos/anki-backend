package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// RegisterAuthRoutes registers authentication routes on the Echo router
func RegisterAuthRoutes(e *echo.Echo, authService primary.IAuthService, jwtService *jwt.JWTService, cacheRepo secondary.ICacheRepository, sessionService primary.ISessionService) {
	authHandler := handlers.NewAuthHandler(authService)
	sessionHandler := handlers.NewSessionHandler(sessionService)

	// Create auth group
	authGroup := e.Group("/api/v1/auth")

	// Create authenticated auth group (requires JWT)
	authMiddleware := middlewares.AuthMiddleware(jwtService, cacheRepo)
	authenticatedAuthGroup := authGroup.Group("", authMiddleware)

	// Register public routes
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.GET("/verify-email", authHandler.VerifyEmail)
	authGroup.POST("/resend-verification", authHandler.ResendVerificationEmail)
	authGroup.POST("/request-password-reset", authHandler.RequestPasswordReset)
	authGroup.POST("/reset-password", authHandler.ResetPassword)

	// Register authenticated routes
	authenticatedAuthGroup.POST("/change-password", authHandler.ChangePassword)

	// Register session management routes
	authenticatedAuthGroup.GET("/sessions", sessionHandler.GetSessions)
	authenticatedAuthGroup.GET("/sessions/:id", sessionHandler.GetSession)
	authenticatedAuthGroup.DELETE("/sessions/:id", sessionHandler.DeleteSession)
	authenticatedAuthGroup.DELETE("/sessions", sessionHandler.DeleteAllSessions)
}
