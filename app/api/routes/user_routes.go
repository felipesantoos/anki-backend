package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// RegisterUserRoutes registers user-related routes (profile, preferences, account)
func RegisterUserRoutes(
	e *echo.Echo,
	userService primary.IUserService,
	profileService primary.IProfileService,
	userPreferencesService primary.IUserPreferencesService,
	jwtService *jwt.JWTService,
	cacheRepo secondary.ICacheRepository,
) {
	userHandler := handlers.NewUserHandler(userService)
	profileHandler := handlers.NewProfileHandler(profileService)
	preferencesHandler := handlers.NewUserPreferencesHandler(userPreferencesService)

	// Auth middleware
	authMiddleware := middlewares.AuthMiddleware(jwtService, cacheRepo)

	// User group
	v1 := e.Group("/api/v1", authMiddleware)

	// Account management
	me := v1.Group("/user/me")
	me.GET("", userHandler.GetMe)
	me.PUT("", userHandler.Update)
	me.DELETE("", userHandler.Delete)

	// Preferences
	prefs := v1.Group("/user/preferences")
	prefs.GET("", preferencesHandler.FindByUserID)
	prefs.PUT("", preferencesHandler.Update)
	prefs.POST("/reset", preferencesHandler.ResetToDefaults)

	// Profiles
	profiles := v1.Group("/profiles")
	profiles.POST("", profileHandler.Create)
	profiles.GET("", profileHandler.FindAll)
	profiles.GET("/:id", profileHandler.FindByID)
	profiles.PUT("/:id", profileHandler.Update)
	profiles.DELETE("/:id", profileHandler.Delete)
	profiles.POST("/:id/sync/enable", profileHandler.EnableSync)
	profiles.POST("/:id/sync/disable", profileHandler.DisableSync)
}

