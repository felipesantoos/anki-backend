package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// RegisterCommunityRoutes registers community-related routes (marketplace, ratings, audit logs)
func RegisterCommunityRoutes(
	e *echo.Echo,
	sharedDeckService primary.ISharedDeckService,
	ratingService primary.ISharedDeckRatingService,
	deletionLogService primary.IDeletionLogService,
	undoHistoryService primary.IUndoHistoryService,
	jwtService *jwt.JWTService,
	cacheRepo secondary.ICacheRepository,
) {
	sharedDeckHandler := handlers.NewSharedDeckHandler(sharedDeckService)
	ratingHandler := handlers.NewSharedDeckRatingHandler(ratingService)
	auditHandler := handlers.NewAuditHandler(deletionLogService, undoHistoryService)

	// Auth middleware
	authMiddleware := middlewares.AuthMiddleware(jwtService, cacheRepo)

	// Marketplace (Public)
	marketplace := e.Group("/api/v1/marketplace")
	marketplace.GET("/decks", sharedDeckHandler.FindAll)
	marketplace.GET("/decks/:id", sharedDeckHandler.FindByID)
	marketplace.GET("/decks/:id/ratings", ratingHandler.FindBySharedDeckID)

	// Marketplace (Auth required)
	authMarketplace := marketplace.Group("", authMiddleware)
	authMarketplace.POST("/decks", sharedDeckHandler.Create)
	authMarketplace.PUT("/decks/:id", sharedDeckHandler.Update)
	authMarketplace.DELETE("/decks/:id", sharedDeckHandler.Delete)
	authMarketplace.POST("/decks/:id/download", sharedDeckHandler.Download)
	authMarketplace.POST("/ratings", ratingHandler.Create)
	authMarketplace.PUT("/decks/:id/ratings", ratingHandler.Update)
	authMarketplace.DELETE("/decks/:id/ratings", ratingHandler.Delete)

	// Audit Logs (Auth required)
	audit := e.Group("/api/v1/audit", authMiddleware)
	audit.GET("/deletions", auditHandler.GetDeletionLogs)
	audit.GET("/undo", auditHandler.GetUndoHistory)
	audit.DELETE("/undo/:id", auditHandler.DeleteUndoHistory)
}

