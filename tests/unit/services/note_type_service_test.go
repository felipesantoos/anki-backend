package services

import (
	"context"
	"testing"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	noteTypeSvc "github.com/felipesantos/anki-backend/core/services/notetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoteTypeService_Create(t *testing.T) {
	mockRepo := new(MockNoteTypeRepository)
	service := noteTypeSvc.NewNoteTypeService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "Basic"
		mockRepo.On("ExistsByName", ctx, userID, name).Return(false, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*notetype.NoteType")).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, "[]", "[]", "{}")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Already Exists", func(t *testing.T) {
		name := "Duplicate"
		mockRepo.On("ExistsByName", ctx, userID, name).Return(true, nil).Once()

		result, err := service.Create(ctx, userID, name, "[]", "[]", "{}")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestNoteTypeService_Update(t *testing.T) {
	mockRepo := new(MockNoteTypeRepository)
	service := noteTypeSvc.NewNoteTypeService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	id := int64(100)

	t.Run("Success", func(t *testing.T) {
		existing, _ := notetype.NewBuilder().WithID(id).WithUserID(userID).WithName("Old Name").Build()
		newName := "New Name"

		mockRepo.On("FindByID", ctx, userID, id).Return(existing, nil).Once()
		mockRepo.On("ExistsByName", ctx, userID, newName).Return(false, nil).Once()
		mockRepo.On("Update", ctx, userID, id, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, userID, id, newName, "[]", "[]", "{}")

		assert.NoError(t, err)
		assert.Equal(t, newName, result.GetName())
		mockRepo.AssertExpectations(t)
	})
}

