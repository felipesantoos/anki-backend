package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
)

// ISharedDeckService defines the interface for shared deck (marketplace) operations
type ISharedDeckService interface {
	// Create publishes a deck to the marketplace
	Create(ctx context.Context, authorID int64, name string, description *string, category *string, packagePath string, packageSize int64, tags []string) (*shareddeck.SharedDeck, error)

	// FindByID finds a shared deck by ID
	FindByID(ctx context.Context, userID int64, id int64) (*shareddeck.SharedDeck, error)

	// FindAll finds all public shared decks with optional filters
	FindAll(ctx context.Context, category *string, tags []string) ([]*shareddeck.SharedDeck, error)

	// Update updates an existing shared deck
	Update(ctx context.Context, authorID int64, id int64, name string, description *string, category *string, isPublic bool, tags []string) (*shareddeck.SharedDeck, error)

	// Delete removes a shared deck from the marketplace (soft delete)
	Delete(ctx context.Context, authorID int64, id int64) error

	// IncrementDownloadCount increments the download counter for a shared deck
	IncrementDownloadCount(ctx context.Context, userID int64, id int64) error
}

