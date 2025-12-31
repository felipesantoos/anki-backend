package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	deckSvc "github.com/felipesantos/anki-backend/core/services/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeckService_Create(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	service := deckSvc.NewDeckService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "New Deck"
		mockRepo.On("Exists", ctx, userID, name, (*int64)(nil)).Return(false, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*deck.Deck")).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, nil, "")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Already Exists", func(t *testing.T) {
		name := "Existing Deck"
		mockRepo.On("Exists", ctx, userID, name, (*int64)(nil)).Return(true, nil).Once()

		result, err := service.Create(ctx, userID, name, nil, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("With Parent Success", func(t *testing.T) {
		name := "Child Deck"
		parentID := int64(10)
		parentDeck, _ := deck.NewBuilder().WithID(parentID).WithUserID(userID).WithName("Parent").Build()

		mockRepo.On("Exists", ctx, userID, name, &parentID).Return(false, nil).Once()
		mockRepo.On("FindByID", ctx, userID, parentID).Return(parentDeck, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, &parentID, "")

		assert.NoError(t, err)
		assert.Equal(t, &parentID, result.GetParentID())
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckService_Delete(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	service := deckSvc.NewDeckService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		deckID := int64(100)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("To Delete").Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()
		mockRepo.On("Delete", ctx, userID, deckID).Return(nil).Once()

		err := service.Delete(ctx, userID, deckID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Prevent Default Deck Deletion", func(t *testing.T) {
		deckID := int64(1)
		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Default").Build()

		mockRepo.On("FindByID", ctx, userID, deckID).Return(d, nil).Once()

		err := service.Delete(ctx, userID, deckID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete the default deck")
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckService_FindByUserID(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	service := deckSvc.NewDeckService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		d1, _ := deck.NewBuilder().WithID(1).WithUserID(userID).WithName("Deck 1").Build()
		d2, _ := deck.NewBuilder().WithID(2).WithUserID(userID).WithName("Deck 2").Build()
		expectedDecks := []*deck.Deck{d1, d2}

		mockRepo.On("FindByUserID", ctx, userID).Return(expectedDecks, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedDecks, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty List", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return([]*deck.Deck{}, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})
}

