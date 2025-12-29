package secondary

import (
	"context"

	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
)

// ISyncMetaRepository defines the interface for sync metadata persistence
// All methods that access specific resources require userID to ensure data isolation
type ISyncMetaRepository interface {
	// Save saves or updates sync metadata in the database
	// If the sync meta has an ID, it updates the existing sync meta
	// If the sync meta has no ID, it creates a new sync meta and returns it with the ID set
	Save(ctx context.Context, userID int64, syncMetaEntity *syncmeta.SyncMeta) error

	// FindByID finds sync metadata by ID, filtering by userID to ensure ownership
	// Returns the sync meta if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*syncmeta.SyncMeta, error)

	// FindByUserID finds all sync metadata for a user
	// Returns a list of sync metadata belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*syncmeta.SyncMeta, error)

	// Update updates existing sync metadata, validating ownership
	// Returns error if sync meta doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, syncMetaEntity *syncmeta.SyncMeta) error

	// Delete deletes sync metadata, validating ownership (hard delete - sync_meta doesn't have soft delete)
	// Returns error if sync meta doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if sync metadata exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByClientID finds sync metadata by client ID for a user
	FindByClientID(ctx context.Context, userID int64, clientID string) (*syncmeta.SyncMeta, error)
}

