package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

// RegisterContentRoutes registers content-related routes (notes, note types)
func RegisterContentRoutes(
	e *echo.Echo,
	noteService primary.INoteService,
	noteTypeService primary.INoteTypeService,
	jwtService *jwt.JWTService,
	cacheRepo secondary.ICacheRepository,
) {
	noteHandler := handlers.NewNoteHandler(noteService)
	noteTypeHandler := handlers.NewNoteTypeHandler(noteTypeService)

	// Auth middleware
	authMiddleware := middlewares.AuthMiddleware(jwtService, cacheRepo)

	// Content group
	v1 := e.Group("/api/v1", authMiddleware)

	// Note Types
	noteTypes := v1.Group("/note-types")
	noteTypes.POST("", noteTypeHandler.Create)
	noteTypes.GET("", noteTypeHandler.FindAll)
	noteTypes.GET("/:id", noteTypeHandler.FindByID)
	noteTypes.PUT("/:id", noteTypeHandler.Update)
	noteTypes.DELETE("/:id", noteTypeHandler.Delete)

	// Notes
	notes := v1.Group("/notes")
	notes.POST("", noteHandler.Create)
	notes.GET("", noteHandler.FindAll)
	notes.GET("/:id", noteHandler.FindByID)
	notes.PUT("/:id", noteHandler.Update)
	notes.DELETE("/:id", noteHandler.Delete)

	// Note Tags
	notes.POST("/:id/tags", noteHandler.AddTag)
	notes.DELETE("/:id/tags/:tag", noteHandler.RemoveTag)
}

