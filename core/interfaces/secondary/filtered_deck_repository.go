package secondary

import (
	"context"

	filtereddeck "github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
)

// IFilteredDeckRepository defines the interface for filtered deck data persistence
// All methods that access specific resources require userID to ensure data isolation
type IFilteredDeckRepository interface {
	// Save saves or updates a filtered deck in the database
	// If the filtered deck has an ID, it updates the existing filtered deck
	// If the filtered deck has no ID, it creates a new filtered deck and returns it with the ID set
	Save(ctx context.Context, userID int64, filteredDeckEntity *filtereddeck.FilteredDeck) error

	// FindByID finds a filtered deck by ID, filtering by userID to ensure ownership
	// Returns the filtered deck if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*filtereddeck.FilteredDeck, error)

	// FindByUserID finds all filtered decks for a user
	// Returns a list of filtered decks belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*filtereddeck.FilteredDeck, error)

	// Update updates an existing filtered deck, validating ownership
	// Returns error if filtered deck doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, filteredDeckEntity *filtereddeck.FilteredDeck) error

	// Delete deletes a filtered deck, validating ownership (soft delete)
	// Returns error if filtered deck doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a filtered deck exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)
}

