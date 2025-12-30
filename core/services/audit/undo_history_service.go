package audit

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// UndoHistoryService implements IUndoHistoryService
type UndoHistoryService struct {
	repo secondary.IUndoHistoryRepository
}

// NewUndoHistoryService creates a new UndoHistoryService instance
func NewUndoHistoryService(repo secondary.IUndoHistoryRepository) primary.IUndoHistoryService {
	return &UndoHistoryService{
		repo: repo,
	}
}

// Create records a new undoable action
func (s *UndoHistoryService) Create(ctx context.Context, userID int64, actionType string, actionData string) (*undohistory.UndoHistory, error) {
	uh, err := undohistory.NewBuilder().
		WithUserID(userID).
		WithOperationType(actionType).
		WithOperationData(actionData).
		WithCreatedAt(time.Now()).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, uh); err != nil {
		return nil, err
	}

	return uh, nil
}

// FindLatest finds the most recent undoable actions for a user
func (s *UndoHistoryService) FindLatest(ctx context.Context, userID int64, limit int) ([]*undohistory.UndoHistory, error) {
	return s.repo.FindLatest(ctx, userID, limit)
}

// Delete removes an undo history record
func (s *UndoHistoryService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

