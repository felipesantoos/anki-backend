package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	auditSvc "github.com/felipesantos/anki-backend/core/services/audit"
	"github.com/felipesantos/anki-backend/pkg/ownership"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeletionLogService_Create(t *testing.T) {
	mockRepo := new(MockDeletionLogRepository)
	mockNoteService := new(MockNoteService)
	mockNoteRepo := new(MockNoteRepository)
	service := auditSvc.NewDeletionLogService(mockRepo, mockNoteService, mockNoteRepo)
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
	mockNoteService := new(MockNoteService)
	mockNoteRepo := new(MockNoteRepository)
	service := auditSvc.NewDeletionLogService(mockRepo, mockNoteService, mockNoteRepo)
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

func TestDeletionLogService_Restore(t *testing.T) {
	mockRepo := new(MockDeletionLogRepository)
	mockNoteService := new(MockNoteService)
	mockNoteRepo := new(MockNoteRepository)
	service := auditSvc.NewDeletionLogService(mockRepo, mockNoteService, mockNoteRepo)
	ctx := context.Background()
	userID := int64(1)
	deletionLogID := int64(100)
	deckID := int64(20)

	// Sample deletion log with valid object_data
	objectDataJSON := `{"guid":"550e8400-e29b-41d4-a716-446655440000","note_type_id":10,"fields":{"Front":"Hello","Back":"World"},"tags":["vocab"]}`
	dl, _ := deletionlog.NewBuilder().
		WithID(deletionLogID).
		WithUserID(userID).
		WithObjectType(deletionlog.ObjectTypeNote).
		WithObjectID(101).
		WithObjectData(objectDataJSON).
		WithDeletedAt(time.Now()).
		Build()

	// Sample restored note
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	restoredNote, _ := note.NewBuilder().
		WithID(201).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(10).
		WithFieldsJSON(`{"Front":"Hello","Back":"World"}`).
		WithTags([]string{"vocab"}).
		Build()

	t.Run("Success - restore note", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dl, nil).Once()
		mockNoteRepo.On("FindByGUID", ctx, userID, "550e8400-e29b-41d4-a716-446655440000").Return(nil, nil).Once()
		mockNoteService.On("Create", ctx, userID, int64(10), deckID, `{"Front":"Hello","Back":"World"}`, []string{"vocab"}).Return(restoredNote, nil).Once()
		mockNoteRepo.On("Update", ctx, userID, int64(201), mock.AnythingOfType("*note.Note")).Return(nil).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, restoredNote.GetID(), result.GetID())
		mockRepo.AssertExpectations(t)
		mockNoteService.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Error - deletion log not found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(nil, ownership.ErrResourceNotFound).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ownership.ErrResourceNotFound))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - deletion log belongs to another user", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(nil, ownership.ErrResourceNotFound).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ownership.ErrResourceNotFound))
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - empty object_data", func(t *testing.T) {
		dlEmpty, _ := deletionlog.NewBuilder().
			WithID(deletionLogID).
			WithUserID(userID).
			WithObjectType(deletionlog.ObjectTypeNote).
			WithObjectID(101).
			WithObjectData("").
			WithDeletedAt(time.Now()).
			Build()
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dlEmpty, nil).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recoverable data")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - invalid object_type", func(t *testing.T) {
		dlCard, _ := deletionlog.NewBuilder().
			WithID(deletionLogID).
			WithUserID(userID).
			WithObjectType(deletionlog.ObjectTypeCard).
			WithObjectID(101).
			WithObjectData(objectDataJSON).
			WithDeletedAt(time.Now()).
			Build()
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dlCard, nil).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only restore notes")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - invalid JSON in object_data", func(t *testing.T) {
		dlInvalid, _ := deletionlog.NewBuilder().
			WithID(deletionLogID).
			WithUserID(userID).
			WithObjectType(deletionlog.ObjectTypeNote).
			WithObjectID(101).
			WithObjectData("invalid json{").
			WithDeletedAt(time.Now()).
			Build()
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dlInvalid, nil).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse object_data JSON")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - missing guid in object_data", func(t *testing.T) {
		invalidDataJSON := `{"note_type_id":10,"fields":{"Front":"Hello"},"tags":[]}`
		dlNoGUID, _ := deletionlog.NewBuilder().
			WithID(deletionLogID).
			WithUserID(userID).
			WithObjectType(deletionlog.ObjectTypeNote).
			WithObjectID(101).
			WithObjectData(invalidDataJSON).
			WithDeletedAt(time.Now()).
			Build()
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dlNoGUID, nil).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing or invalid guid")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - note already restored (GUID conflict)", func(t *testing.T) {
		existingNote, _ := note.NewBuilder().
			WithID(301).
			WithUserID(userID).
			WithGUID(guid).
			WithNoteTypeID(10).
			WithFieldsJSON(`{"Front":"Existing"}`).
			WithTags([]string{}).
			Build()
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dl, nil).Once()
		mockNoteRepo.On("FindByGUID", ctx, userID, "550e8400-e29b-41d4-a716-446655440000").Return(existingNote, nil).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists and is not deleted")
		assert.Contains(t, err.Error(), "already restored")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Error - note service Create fails", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dl, nil).Once()
		mockNoteRepo.On("FindByGUID", ctx, userID, "550e8400-e29b-41d4-a716-446655440000").Return(nil, nil).Once()
		mockNoteService.On("Create", ctx, userID, int64(10), deckID, `{"Front":"Hello","Back":"World"}`, []string{"vocab"}).Return(nil, errors.New("note type not found")).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create note")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockNoteService.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Error - Update GUID fails", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, deletionLogID).Return(dl, nil).Once()
		mockNoteRepo.On("FindByGUID", ctx, userID, "550e8400-e29b-41d4-a716-446655440000").Return(nil, nil).Once()
		mockNoteService.On("Create", ctx, userID, int64(10), deckID, `{"Front":"Hello","Back":"World"}`, []string{"vocab"}).Return(restoredNote, nil).Once()
		mockNoteRepo.On("Update", ctx, userID, int64(201), mock.AnythingOfType("*note.Note")).Return(errors.New("update failed")).Once()

		result, err := service.Restore(ctx, userID, deletionLogID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set original GUID")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockNoteService.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
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

