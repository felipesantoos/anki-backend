package sync

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// SyncMetaService implements ISyncMetaService
type SyncMetaService struct {
	repo secondary.ISyncMetaRepository
}

// NewSyncMetaService creates a new SyncMetaService instance
func NewSyncMetaService(repo secondary.ISyncMetaRepository) primary.ISyncMetaService {
	return &SyncMetaService{
		repo: repo,
	}
}

// FindByUserID finds sync metadata for a user
func (s *SyncMetaService) FindByUserID(ctx context.Context, userID int64) (*syncmeta.SyncMeta, error) {
	metas, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(metas) == 0 {
		return nil, nil
	}
	return metas[0], nil
}

// Update updates sync metadata after a synchronization
func (s *SyncMetaService) Update(ctx context.Context, userID int64, clientID string, lastSyncUSN int64) (*syncmeta.SyncMeta, error) {
	existing, err := s.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing == nil {
		// Create new if not exists
		newMeta, err := syncmeta.NewBuilder().
			WithUserID(userID).
			WithClientID(clientID).
			WithLastSync(now).
			WithLastSyncUSN(lastSyncUSN).
			WithCreatedAt(now).
			WithUpdatedAt(now).
			Build()
		
		if err != nil {
			return nil, err
		}

		if err := s.repo.Save(ctx, userID, newMeta); err != nil {
			return nil, err
		}
		return newMeta, nil
	} else {
		// Update existing
		existing.SetClientID(clientID)
		existing.SetLastSync(now)
		existing.SetLastSyncUSN(lastSyncUSN)
		existing.SetUpdatedAt(now)
		if err := s.repo.Update(ctx, userID, existing.GetID(), existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
}

