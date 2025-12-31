package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
)

// IDeckRepository defines the interface for deck data persistence
// All methods that access specific resources require userID to ensure data isolation
type IDeckRepository interface {
	// CreateDefaultDeck creates a default deck for a user
	// The default deck is created with the name "Default" and standard configuration
	// Returns the deck ID or an error if creation fails
	CreateDefaultDeck(ctx context.Context, userID int64) (int64, error)

	// FindByID finds a deck by ID, filtering by userID to ensure ownership
	// Returns the deck if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, deckID int64) (*deck.Deck, error)

	// FindByUserID finds all decks for a user
	// Returns a list of decks belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*deck.Deck, error)

	// FindByParentID finds all decks with a specific parent ID, filtering by userID
	// Returns a list of decks belonging to the user with the specified parent
	FindByParentID(ctx context.Context, userID int64, parentID int64) ([]*deck.Deck, error)

	// Save creates or updates a deck
	// For updates, validates that the deck belongs to the user
	Save(ctx context.Context, userID int64, deckEntity *deck.Deck) error

	// Update updates an existing deck, validating ownership
	// Returns error if deck doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, deckID int64, deckEntity *deck.Deck) error

	// Delete deletes a deck, validating ownership
	// Returns error if deck doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, deckID int64) error

	// Exists checks if a deck with the given name exists for the user at the specified parent level
	Exists(ctx context.Context, userID int64, name string, parentID *int64) (bool, error)

	// GetStats retrieves study statistics for a deck
	GetStats(ctx context.Context, userID int64, deckID int64) (*deck.DeckStats, error)
}
