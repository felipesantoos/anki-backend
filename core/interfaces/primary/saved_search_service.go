package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
)

// ISavedSearchService defines the interface for saved search management
type ISavedSearchService interface {
	// Create creates a new saved search
	Create(ctx context.Context, userID int64, name string, query string) (*savedsearch.SavedSearch, error)

	// FindByUserID finds all saved searches for a user
	FindByUserID(ctx context.Context, userID int64) ([]*savedsearch.SavedSearch, error)

	// Update updates an existing saved search
	Update(ctx context.Context, userID int64, id int64, name string, query string) (*savedsearch.SavedSearch, error)

	// Delete deletes a saved search
	Delete(ctx context.Context, userID int64, id int64) error
}

