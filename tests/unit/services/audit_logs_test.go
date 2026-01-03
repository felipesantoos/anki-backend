package services

import (
	"context"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
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

func TestDeletionLogService_FindRecent(t *testing.T) {
	mockRepo := new(MockDeletionLogRepository)
	service := auditSvc.NewDeletionLogService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	// Sample deletion logs for testing
	dl1, _ := deletionlog.NewBuilder().
		WithID(1).
		WithUserID(userID).
		WithObjectType("note").
		WithObjectID(100).
		WithObjectData(`{"id":100}`).
		WithDeletedAt(time.Now()).
		Build()
	dl2, _ := deletionlog.NewBuilder().
		WithID(2).
		WithUserID(userID).
		WithObjectType("card").
		WithObjectID(200).
		WithObjectData(`{"id":200}`).
		WithDeletedAt(time.Now()).
		Build()

	t.Run("Success - with default parameters", func(t *testing.T) {
		// When limit and days are 0 or negative, defaults should be applied
		mockRepo.On("FindRecent", ctx, userID, 20, 7).Return([]*deletionlog.DeletionLog{dl1, dl2}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 0, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - with custom valid parameters", func(t *testing.T) {
		mockRepo.On("FindRecent", ctx, userID, 10, 5).Return([]*deletionlog.DeletionLog{dl1}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 10, 5)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, dl1.GetID(), result[0].GetID())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - limit exceeds maximum, should be capped at 100", func(t *testing.T) {
		// Service should cap limit at 100
		mockRepo.On("FindRecent", ctx, userID, 100, 7).Return([]*deletionlog.DeletionLog{dl1}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 150, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - days exceeds maximum, should be capped at 365", func(t *testing.T) {
		// Service should cap days at 365
		mockRepo.On("FindRecent", ctx, userID, 20, 365).Return([]*deletionlog.DeletionLog{dl1}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 0, 500)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - negative limit, should use default", func(t *testing.T) {
		// Negative limit should default to 20
		mockRepo.On("FindRecent", ctx, userID, 20, 7).Return([]*deletionlog.DeletionLog{}, nil).Once()

		result, err := service.FindRecent(ctx, userID, -5, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - negative days, should use default", func(t *testing.T) {
		// Negative days should default to 7
		mockRepo.On("FindRecent", ctx, userID, 20, 7).Return([]*deletionlog.DeletionLog{}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 0, -10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - boundary values: limit = 1", func(t *testing.T) {
		mockRepo.On("FindRecent", ctx, userID, 1, 7).Return([]*deletionlog.DeletionLog{dl1}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 1, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - boundary values: limit = 100", func(t *testing.T) {
		mockRepo.On("FindRecent", ctx, userID, 100, 7).Return([]*deletionlog.DeletionLog{dl1, dl2}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 100, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - boundary values: days = 1", func(t *testing.T) {
		mockRepo.On("FindRecent", ctx, userID, 20, 1).Return([]*deletionlog.DeletionLog{dl1}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 0, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - boundary values: days = 365", func(t *testing.T) {
		mockRepo.On("FindRecent", ctx, userID, 20, 365).Return([]*deletionlog.DeletionLog{dl1, dl2}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 0, 365)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - empty result", func(t *testing.T) {
		mockRepo.On("FindRecent", ctx, userID, 20, 7).Return([]*deletionlog.DeletionLog{}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 0, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - repository returns error", func(t *testing.T) {
		expectedError := assert.AnError
		mockRepo.On("FindRecent", ctx, userID, 20, 7).Return(nil, expectedError).Once()

		result, err := service.FindRecent(ctx, userID, 0, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - both parameters exceed maximum, both should be capped", func(t *testing.T) {
		// Both limit and days exceed maximum, both should be capped
		mockRepo.On("FindRecent", ctx, userID, 100, 365).Return([]*deletionlog.DeletionLog{dl1}, nil).Once()

		result, err := service.FindRecent(ctx, userID, 200, 500)

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

