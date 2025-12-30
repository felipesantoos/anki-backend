package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
)

// IFilteredDeckService defines the interface for filtered deck management
type IFilteredDeckService interface {
	// Create creates a new filtered deck
	Create(ctx context.Context, userID int64, name string, searchFilter string, limit int, orderBy string, reschedule bool) (*filtereddeck.FilteredDeck, error)

	// FindByUserID finds all filtered decks for a user
	FindByUserID(ctx context.Context, userID int64) ([]*filtereddeck.FilteredDeck, error)

	// Update updates an existing filtered deck
	Update(ctx context.Context, userID int64, id int64, name string, searchFilter string, limit int, orderBy string, reschedule bool) (*filtereddeck.FilteredDeck, error)

	// Delete deletes a filtered deck
	Delete(ctx context.Context, userID int64, id int64) error
}

