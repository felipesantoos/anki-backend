package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	"github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	auditSvc "github.com/felipesantos/anki-backend/core/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeletionLogService_Create(t *testing.T) {
	mockRepo := new(MockDeletionLogRepository)
	service := auditSvc.NewDeletionLogService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, "deck", 100)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUndoHistoryService_Create(t *testing.T) {
	mockRepo := new(MockUndoHistoryRepository)
	service := auditSvc.NewUndoHistoryService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, undohistory.OperationTypeEditNote, "{\"id\":100}")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCheckDatabaseLogService_Create(t *testing.T) {
	mockRepo := new(MockCheckDatabaseLogRepository)
	service := auditSvc.NewCheckDatabaseLogService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, checkdatabaselog.CheckStatusCompleted, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

