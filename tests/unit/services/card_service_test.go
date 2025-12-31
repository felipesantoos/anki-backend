package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	cardSvc "github.com/felipesantos/anki-backend/core/services/card"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCardService_Suspend(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := cardSvc.NewCardService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	cardID := int64(100)

	t.Run("Success", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithNoteID(1).WithDeckID(1).Build()
		assert.False(t, c.GetSuspended())

		mockRepo.On("FindByID", ctx, userID, cardID).Return(c, nil).Once()
		mockRepo.On("Update", ctx, userID, cardID, mock.Anything).Return(nil).Once()

		err := service.Suspend(ctx, userID, cardID)

		assert.NoError(t, err)
		assert.True(t, c.GetSuspended())
		mockRepo.AssertExpectations(t)
	})
}

func TestCardService_SetFlag(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := cardSvc.NewCardService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	cardID := int64(100)

	t.Run("Success", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithNoteID(1).WithDeckID(1).Build()
		flag := 3

		mockRepo.On("FindByID", ctx, userID, cardID).Return(c, nil).Once()
		mockRepo.On("Update", ctx, userID, cardID, mock.Anything).Return(nil).Once()

		err := service.SetFlag(ctx, userID, cardID, flag)

		assert.NoError(t, err)
		assert.Equal(t, flag, c.GetFlag())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Flag", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithNoteID(1).WithDeckID(1).Build()
		flag := 9 // Invalid

		mockRepo.On("FindByID", ctx, userID, cardID).Return(c, nil).Once()

		err := service.SetFlag(ctx, userID, cardID, flag)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCardService_CountByDeckAndState(t *testing.T) {
	mockRepo := new(MockCardRepository)
	service := cardSvc.NewCardService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success New", func(t *testing.T) {
		mockRepo.On("CountByDeckAndState", ctx, userID, deckID, valueobjects.CardStateNew).Return(5, nil).Once()

		count, err := service.CountByDeckAndState(ctx, userID, deckID, "new")

		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid State", func(t *testing.T) {
		count, err := service.CountByDeckAndState(ctx, userID, deckID, "invalid")

		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "invalid card state")
	})
}

