package routes

import (
	"github.com/felipesantos/anki-backend/app/api/handlers"
	"github.com/felipesantos/anki-backend/app/api/middlewares"
	"github.com/felipesantos/anki-backend/dicontainer"
)

// RegisterStudyRoutes registers study-related routes (decks, cards, reviews)
func (r *Router) RegisterStudyRoutes() {
	deckService := dicontainer.GetDeckService()
	presetService := dicontainer.GetDeckOptionsPresetService()
	deckStatsService := dicontainer.GetDeckStatsService()
	filteredDeckService := dicontainer.GetFilteredDeckService()
	cardService := dicontainer.GetCardService()
	reviewService := dicontainer.GetReviewService()

	deckHandler := handlers.NewDeckHandler(deckService)
	presetHandler := handlers.NewDeckOptionsPresetHandler(presetService)
	deckStatsHandler := handlers.NewDeckStatsHandler(deckStatsService)
	filteredDeckHandler := handlers.NewFilteredDeckHandler(filteredDeckService)
	cardHandler := handlers.NewCardHandler(cardService)
	reviewHandler := handlers.NewReviewHandler(reviewService)

	// Auth middleware
	authMiddleware := middlewares.AuthMiddleware(r.jwtSvc, r.rdb)

	// Study group
	v1 := r.echo.Group("/api/v1", authMiddleware)

	// Decks
	decks := v1.Group("/decks")
	decks.POST("", deckHandler.Create)
	decks.GET("", deckHandler.FindAll)
	decks.GET("/:id", deckHandler.FindByID)
	decks.GET("/:id/stats", deckStatsHandler.GetStats)
	decks.GET("/:id/options", deckHandler.GetOptions)
	decks.PUT("/:id/options", deckHandler.UpdateOptions)
	decks.PUT("/:id", deckHandler.Update)
	decks.DELETE("/:id", deckHandler.Delete)

	// Deck Options Presets
	presets := v1.Group("/deck-options-presets")
	presets.POST("", presetHandler.Create)
	presets.GET("", presetHandler.FindAll)
	presets.PUT("/:id", presetHandler.Update)
	presets.DELETE("/:id", presetHandler.Delete)
	presets.POST("/:id/apply", presetHandler.ApplyToDecks)

	// Filtered Decks
	filteredDecks := v1.Group("/filtered-decks")
	filteredDecks.POST("", filteredDeckHandler.Create)
	filteredDecks.GET("", filteredDeckHandler.FindAll)
	filteredDecks.PUT("/:id", filteredDeckHandler.Update)
	filteredDecks.DELETE("/:id", filteredDeckHandler.Delete)

	// Cards (via Decks)
	decks.GET("/:deckID/cards", cardHandler.FindByDeckID)
	decks.GET("/:deckID/cards/due", cardHandler.FindDueCards)

	// Cards (Direct)
	cards := v1.Group("/cards")
	cards.GET("", cardHandler.FindAll)
	cards.GET("/:id", cardHandler.FindByID)
	cards.POST("/:id/suspend", cardHandler.Suspend)
	cards.POST("/:id/unsuspend", cardHandler.Unsuspend)
	cards.POST("/:id/bury", cardHandler.Bury)
	cards.POST("/:id/unbury", cardHandler.Unbury)
	cards.POST("/:id/flag", cardHandler.SetFlag)
	cards.DELETE("/:id", cardHandler.Delete)

	// Reviews
	reviews := v1.Group("/reviews")
	reviews.POST("", reviewHandler.Create)
	
	// Card Reviews
	cards.GET("/:cardID/reviews", reviewHandler.FindByCardID)
}

