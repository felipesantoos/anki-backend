package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
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
}

