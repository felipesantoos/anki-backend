package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
)

// IDeckService defines the interface for deck management operations
type IDeckService interface {
	// Create creates a new deck for a user
	Create(ctx context.Context, userID int64, name string, parentID *int64, optionsJSON string) (*deck.Deck, error)

	// FindByID finds a deck by ID, validating ownership
	FindByID(ctx context.Context, userID int64, id int64) (*deck.Deck, error)

	// FindByUserID finds all decks for a user
	FindByUserID(ctx context.Context, userID int64) ([]*deck.Deck, error)

	// Update updates an existing deck
	Update(ctx context.Context, userID int64, id int64, name string, parentID *int64, optionsJSON string) (*deck.Deck, error)

	// Delete deletes a deck (soft delete)
	Delete(ctx context.Context, userID int64, id int64) error

	// CreateDefaultDeck creates the initial "Default" deck for a user
	CreateDefaultDeck(ctx context.Context, userID int64) (*deck.Deck, error)
}

