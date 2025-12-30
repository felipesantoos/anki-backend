package services

import (
	"context"
	"testing"

	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	ratingSvc "github.com/felipesantos/anki-backend/core/services/shareddeckrating"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSharedDeckRatingService_Create(t *testing.T) {
	mockRepo := new(MockSharedDeckRatingRepository)
	service := ratingSvc.NewSharedDeckRatingService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	deckID := int64(100)

	t.Run("Success", func(t *testing.T) {
		ratingVO, _ := valueobjects.NewSharedDeckRating(5)
		mockRepo.On("FindByUserIDAndSharedDeckID", ctx, userID, deckID).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*shareddeckrating.SharedDeckRating")).Return(nil).Once()

		result, err := service.Create(ctx, userID, deckID, 5, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, ratingVO, result.GetRating())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Already Rated", func(t *testing.T) {
		existing, _ := shareddeckrating.NewBuilder().WithID(1).WithUserID(userID).WithSharedDeckID(deckID).Build()
		mockRepo.On("FindByUserIDAndSharedDeckID", ctx, userID, deckID).Return(existing, nil).Once()

		result, err := service.Create(ctx, userID, deckID, 5, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already rated")
		mockRepo.AssertExpectations(t)
	})
}

