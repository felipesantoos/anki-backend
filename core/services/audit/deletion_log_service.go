package audit

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// DeletionLogService implements IDeletionLogService
type DeletionLogService struct {
	repo secondary.IDeletionLogRepository
}

// NewDeletionLogService creates a new DeletionLogService instance
func NewDeletionLogService(repo secondary.IDeletionLogRepository) primary.IDeletionLogService {
	return &DeletionLogService{
		repo: repo,
	}
}

// Create records a new deletion event
func (s *DeletionLogService) Create(ctx context.Context, userID int64, objectType string, objectID int64) (*deletionlog.DeletionLog, error) {
	dl, err := deletionlog.NewBuilder().
		WithUserID(userID).
		WithObjectType(objectType).
		WithObjectID(objectID).
		WithDeletedAt(time.Now()).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, dl); err != nil {
		return nil, err
	}

	return dl, nil
}

// FindByUserID finds deletion logs for a user
func (s *DeletionLogService) FindByUserID(ctx context.Context, userID int64) ([]*deletionlog.DeletionLog, error) {
	return s.repo.FindByUserID(ctx, userID)
}

