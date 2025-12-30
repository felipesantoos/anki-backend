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
}

