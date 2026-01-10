package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
)

// ICardService defines the interface for card management operations
type ICardService interface {
	// FindByID finds a card by ID
	FindByID(ctx context.Context, userID int64, id int64) (*card.Card, error)

	// FindByDeckID finds all cards in a deck
	FindByDeckID(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error)

	// Update updates an existing card
	Update(ctx context.Context, userID int64, cardEntity *card.Card) error

	// Delete deletes a card
	Delete(ctx context.Context, userID int64, id int64) error

	// Suspend suspends a card
	Suspend(ctx context.Context, userID int64, id int64) error

	// Unsuspend unsuspends a card
	Unsuspend(ctx context.Context, userID int64, id int64) error

	// Bury buries a card
	Bury(ctx context.Context, userID int64, id int64) error

	// Unbury unburies a card
	Unbury(ctx context.Context, userID int64, id int64) error

	// SetFlag sets a colored flag on a card
	SetFlag(ctx context.Context, userID int64, id int64, flag int) error

	// FindDueCards finds cards that are due for review in a deck
	FindDueCards(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error)

	// CountByDeckAndState counts cards with a specific state in a deck
	CountByDeckAndState(ctx context.Context, userID int64, deckID int64, state string) (int, error)

	// FindAll finds cards for a user based on filters and pagination
	// Returns: list of cards, total count (for pagination), error
	FindAll(ctx context.Context, userID int64, filters card.CardFilters) ([]*card.Card, int, error)

	// GetInfo returns detailed card information including note data, deck/note type names, and review history
	GetInfo(ctx context.Context, userID int64, cardID int64) (*card.CardInfo, error)

	// Reset resets a card (type can be "new" or "forget")
	Reset(ctx context.Context, userID int64, id int64, resetType string) error

	// SetDueDate manually sets the due date for a card
	SetDueDate(ctx context.Context, userID int64, id int64, due int64) error

	// FindLeeches finds cards that are difficult to memorize (leeches)
	FindLeeches(ctx context.Context, userID int64, limit, offset int) ([]*card.Card, int, error)

	// Reposition changes the order new cards will appear in
	Reposition(ctx context.Context, userID int64, cardIDs []int64, start int, step int, shift bool) error

	// GetPosition returns the ordinal position of a card
	GetPosition(ctx context.Context, userID int64, cardID int64) (int, error)
}
