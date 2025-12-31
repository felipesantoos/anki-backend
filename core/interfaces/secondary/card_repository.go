package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// ICardRepository defines the interface for card data persistence
// All methods that access specific resources require userID to ensure data isolation
// Ownership is validated via JOIN with decks table (cards belong to decks, decks belong to users)
type ICardRepository interface {
	// Save saves or updates a card in the database
	// If the card has an ID, it updates the existing card
	// If the card has no ID, it creates a new card and returns it with the ID set
	// Validates ownership via deck_id
	Save(ctx context.Context, userID int64, cardEntity *card.Card) error

	// FindByID finds a card by ID, filtering by userID via deck ownership to ensure ownership
	// Returns the card if found and belongs to user's deck, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*card.Card, error)

	// FindByDeckID finds all cards in a deck, validating deck ownership
	// Returns a list of cards in the deck
	FindByDeckID(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error)

	// Update updates an existing card, validating ownership via deck
	// Returns error if card doesn't exist or doesn't belong to user's deck
	Update(ctx context.Context, userID int64, id int64, cardEntity *card.Card) error

	// Delete deletes a card, validating ownership via deck (hard delete - cards don't have soft delete)
	// Returns error if card doesn't exist or doesn't belong to user's deck
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a card exists and belongs to a user's deck
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByNoteID finds all cards generated from a specific note, validating ownership
	FindByNoteID(ctx context.Context, userID int64, noteID int64) ([]*card.Card, error)

	// FindDueCards finds cards that are due for review in a deck
	// dueTimestamp is the current timestamp in milliseconds
	FindDueCards(ctx context.Context, userID int64, deckID int64, dueTimestamp int64) ([]*card.Card, error)

	// FindByState finds all cards with a specific state in a deck
	FindByState(ctx context.Context, userID int64, deckID int64, state valueobjects.CardState) ([]*card.Card, error)

	// CountByDeckAndState counts cards with a specific state in a deck
	CountByDeckAndState(ctx context.Context, userID int64, deckID int64, state valueobjects.CardState) (int, error)

	// MoveCards moves all cards from a source deck (including sub-decks) to a target deck
	MoveCards(ctx context.Context, userID int64, srcDeckID int64, targetDeckID int64) error

	// DeleteByDeckRecursive deletes all cards from a deck and its sub-decks
	DeleteByDeckRecursive(ctx context.Context, userID int64, deckID int64) error
}

