package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	sharedDeckSvc "github.com/felipesantos/anki-backend/core/services/shareddeck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSharedDeckService_Create(t *testing.T) {
	mockRepo := new(MockSharedDeckRepository)
	service := sharedDeckSvc.NewSharedDeckService(mockRepo)
	ctx := context.Background()
	authorID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "Japanese Core 2k"
		mockRepo.On("Save", ctx, authorID, mock.AnythingOfType("*shareddeck.SharedDeck")).Return(nil).Once()

		result, err := service.Create(ctx, authorID, name, nil, nil, "/path/pkg.apkg", 1024, []string{"lang"})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})
}

func TestSharedDeckService_Update(t *testing.T) {
	mockRepo := new(MockSharedDeckRepository)
	service := sharedDeckSvc.NewSharedDeckService(mockRepo)
	ctx := context.Background()
	authorID := int64(1)
	id := int64(100)

	t.Run("Success", func(t *testing.T) {
		existing, _ := shareddeck.NewBuilder().WithID(id).WithAuthorID(authorID).WithName("Old").Build()
		newName := "New Name"

		mockRepo.On("FindByID", ctx, authorID, id).Return(existing, nil).Once()
		mockRepo.On("Update", ctx, authorID, id, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, authorID, id, newName, nil, nil, true, nil)

		assert.NoError(t, err)
		assert.Equal(t, newName, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		existing, _ := shareddeck.NewBuilder().WithID(id).WithAuthorID(999).WithName("Not Mine").Build()

		mockRepo.On("FindByID", ctx, authorID, id).Return(existing, nil).Once()

		result, err := service.Update(ctx, authorID, id, "Hack", nil, nil, true, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "access denied")
		mockRepo.AssertExpectations(t)
	})
}

