package routes

import (
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/dicontainer"
)

// RegisterSearchRoutes registers search-related routes
func (r *Router) RegisterSearchRoutes() {
	searchService := dicontainer.GetSearchService()
	searchHandler := handlers.NewSearchHandler(searchService)

	// Auth middleware
	authMiddleware := middlewares.AuthMiddleware(r.jwtSvc, r.rdb)

	// Search group
	v1 := r.echo.Group("/api/v1", authMiddleware)

	// Advanced search
	search := v1.Group("/search")
	search.POST("/advanced", searchHandler.SearchAdvanced)
}

