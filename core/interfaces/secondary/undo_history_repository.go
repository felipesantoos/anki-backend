package secondary

import (
	"context"

	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
)

// IUndoHistoryRepository defines the interface for undo history data persistence
// All methods that access specific resources require userID to ensure data isolation
type IUndoHistoryRepository interface {
	// Save saves or updates an undo history entry in the database
	// If the undo history has an ID, it updates the existing entry
	// If the undo history has no ID, it creates a new entry and returns it with the ID set
	Save(ctx context.Context, userID int64, undoHistoryEntity *undohistory.UndoHistory) error

	// FindByID finds an undo history entry by ID, filtering by userID to ensure ownership
	// Returns the undo history if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*undohistory.UndoHistory, error)

	// FindByUserID finds all undo history entries for a user
	// Returns a list of undo history entries belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*undohistory.UndoHistory, error)

	// Update updates an existing undo history entry, validating ownership
	// Returns error if undo history doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, undoHistoryEntity *undohistory.UndoHistory) error

	// Delete deletes an undo history entry, validating ownership (hard delete - undo_history doesn't have soft delete)
	// Returns error if undo history doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if an undo history entry exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindLatest finds the latest undo history entries for a user
	FindLatest(ctx context.Context, userID int64, limit int) ([]*undohistory.UndoHistory, error)
}

