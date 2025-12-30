package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
)

// IUndoHistoryService defines the interface for undo operations history
type IUndoHistoryService interface {
	// Create records a new undoable action
	Create(ctx context.Context, userID int64, actionType string, actionData string) (*undohistory.UndoHistory, error)

	// FindLatest finds the most recent undoable actions for a user
	FindLatest(ctx context.Context, userID int64, limit int) ([]*undohistory.UndoHistory, error)

	// Delete removes an undo history record
	Delete(ctx context.Context, userID int64, id int64) error
}

