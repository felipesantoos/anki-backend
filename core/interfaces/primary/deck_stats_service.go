package primary

import (
	"context"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
)

// IDeckStatsService defines the interface for deck statistics operations
type IDeckStatsService interface {
	// GetStats retrieves study statistics for a specific deck
	GetStats(ctx context.Context, userID int64, deckID int64) (*deck.DeckStats, error)
}

