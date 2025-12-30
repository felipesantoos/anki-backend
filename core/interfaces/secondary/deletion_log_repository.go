package secondary

import (
	"context"

	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
)

// IDeletionLogRepository defines the interface for deletion log data persistence
// All methods that access specific resources require userID to ensure data isolation
type IDeletionLogRepository interface {
	// Save saves or updates a deletion log in the database
	// If the deletion log has an ID, it updates the existing deletion log
	// If the deletion log has no ID, it creates a new deletion log and returns it with the ID set
	Save(ctx context.Context, userID int64, deletionLogEntity *deletionlog.DeletionLog) error

	// FindByID finds a deletion log by ID, filtering by userID to ensure ownership
	// Returns the deletion log if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*deletionlog.DeletionLog, error)

	// FindByUserID finds all deletion logs for a user
	// Returns a list of deletion logs belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*deletionlog.DeletionLog, error)

	// Update updates an existing deletion log, validating ownership
	// Returns error if deletion log doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, deletionLogEntity *deletionlog.DeletionLog) error

	// Delete deletes a deletion log, validating ownership (hard delete - deletions_log doesn't have soft delete)
	// Returns error if deletion log doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a deletion log exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByObjectType finds all deletion logs of a specific object type for a user
	FindByObjectType(ctx context.Context, userID int64, objectType string) ([]*deletionlog.DeletionLog, error)
}

