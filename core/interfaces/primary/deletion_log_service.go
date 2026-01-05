package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
)

// IDeletionLogService defines the interface for deletion audit logging
type IDeletionLogService interface {
	// Create records a new deletion event
	Create(ctx context.Context, userID int64, objectType string, objectID int64) (*deletionlog.DeletionLog, error)

	// FindByUserID finds deletion logs for a user
	FindByUserID(ctx context.Context, userID int64) ([]*deletionlog.DeletionLog, error)

	// FindRecent finds recent deletion logs for a user within a specified time period
	// limit: maximum number of records to return (default: 20, max: 100)
	// days: number of days to look back (default: 7, max: 365)
	// Returns deletion logs ordered by deleted_at DESC, limited to the specified count
	FindRecent(ctx context.Context, userID int64, limit int, days int) ([]*deletionlog.DeletionLog, error)

	// Restore restores a deleted note from a deletion log entry
	// Validates ownership, parses object_data JSON, validates note type and deck,
	// handles GUID conflicts, and creates the note with cards
	// Returns the restored note or an error
	Restore(ctx context.Context, userID int64, deletionLogID int64, deckID int64) (*note.Note, error)
}

