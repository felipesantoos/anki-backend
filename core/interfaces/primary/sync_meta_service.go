package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
)

// ISyncMetaService defines the interface for synchronization metadata management
type ISyncMetaService interface {
	// FindByUserID finds sync metadata for a user
	FindByUserID(ctx context.Context, userID int64) (*syncmeta.SyncMeta, error)

	// Update updates sync metadata after a synchronization
	Update(ctx context.Context, userID int64, clientID string, lastSyncUSN int64) (*syncmeta.SyncMeta, error)
}

