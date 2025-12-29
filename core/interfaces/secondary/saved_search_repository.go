package secondary

import (
	"context"

	savedsearch "github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
)

// ISavedSearchRepository defines the interface for saved search data persistence
// All methods that access specific resources require userID to ensure data isolation
type ISavedSearchRepository interface {
	// Save saves or updates a saved search in the database
	// If the saved search has an ID, it updates the existing saved search
	// If the saved search has no ID, it creates a new saved search and returns it with the ID set
	Save(ctx context.Context, userID int64, savedSearchEntity *savedsearch.SavedSearch) error

	// FindByID finds a saved search by ID, filtering by userID to ensure ownership
	// Returns the saved search if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*savedsearch.SavedSearch, error)

	// FindByUserID finds all saved searches for a user
	// Returns a list of saved searches belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*savedsearch.SavedSearch, error)

	// Update updates an existing saved search, validating ownership
	// Returns error if saved search doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, savedSearchEntity *savedsearch.SavedSearch) error

	// Delete deletes a saved search, validating ownership (soft delete)
	// Returns error if saved search doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a saved search exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)
}

