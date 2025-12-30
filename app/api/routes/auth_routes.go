package routes

import (
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/dicontainer"
)

// RegisterAuthRoutes registers authentication routes on the Router
func (r *Router) RegisterAuthRoutes() {
	authService := dicontainer.GetAuthService()
	sessionService := dicontainer.GetSessionService()
	
	authHandler := handlers.NewAuthHandler(authService)
	sessionHandler := handlers.NewSessionHandler(sessionService)

	// Create auth group
	authGroup := r.echo.Group("/api/v1/auth")

	// Create authenticated auth group (requires JWT)
	authMiddleware := middlewares.AuthMiddleware(r.jwtSvc, r.rdb)
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
