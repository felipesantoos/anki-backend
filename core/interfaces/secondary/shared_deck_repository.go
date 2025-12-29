package secondary

import (
	"context"

	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
)

// ISharedDeckRepository defines the interface for shared deck data persistence
// Shared decks can be public (visible to all) or private (only to author)
type ISharedDeckRepository interface {
	// Save saves or updates a shared deck in the database
	// If the shared deck has an ID, it updates the existing shared deck
	// If the shared deck has no ID, it creates a new shared deck and returns it with the ID set
	Save(ctx context.Context, authorID int64, sharedDeckEntity *shareddeck.SharedDeck) error

	// FindByID finds a shared deck by ID
	// For non-authors, only returns public shared decks
	// For authors, returns their own shared decks regardless of public status
	FindByID(ctx context.Context, userID int64, id int64) (*shareddeck.SharedDeck, error)

	// FindByAuthorID finds all shared decks by an author
	// Only returns decks belonging to the author (for ownership validation)
	FindByAuthorID(ctx context.Context, authorID int64) ([]*shareddeck.SharedDeck, error)

	// Update updates an existing shared deck, validating ownership
	// Returns error if shared deck doesn't exist or doesn't belong to user
	Update(ctx context.Context, authorID int64, id int64, sharedDeckEntity *shareddeck.SharedDeck) error

	// Delete deletes a shared deck, validating ownership (soft delete)
	// Returns error if shared deck doesn't exist or doesn't belong to user
	Delete(ctx context.Context, authorID int64, id int64) error

	// Exists checks if a shared deck exists
	Exists(ctx context.Context, id int64) (bool, error)

	// Specific methods
	// FindPublic finds all public shared decks (visible to all users)
	FindPublic(ctx context.Context, limit, offset int) ([]*shareddeck.SharedDeck, error)

	// FindByCategory finds public shared decks by category
	FindByCategory(ctx context.Context, category string, limit, offset int) ([]*shareddeck.SharedDeck, error)

	// FindFeatured finds featured public shared decks
	FindFeatured(ctx context.Context, limit int) ([]*shareddeck.SharedDeck, error)
}

