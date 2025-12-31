package deck

import (
	"context"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// DeckStatsService implements IDeckStatsService
type DeckStatsService struct {
	deckRepo secondary.IDeckRepository
}

// NewDeckStatsService creates a new DeckStatsService instance
func NewDeckStatsService(deckRepo secondary.IDeckRepository) primary.IDeckStatsService {
	return &DeckStatsService{
		deckRepo: deckRepo,
	}
}

// GetStats retrieves study statistics for a specific deck
func (s *DeckStatsService) GetStats(ctx context.Context, userID int64, deckID int64) (*deck.DeckStats, error) {
	return s.deckRepo.GetStats(ctx, userID, deckID)
}

