package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	reviewSvc "github.com/felipesantos/anki-backend/core/services/review"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReviewService_Create(t *testing.T) {
	mockReviewRepo := new(MockReviewRepository)
	mockCardRepo := new(MockCardRepository)
	mockTM := new(MockTransactionManager)
	service := reviewSvc.NewReviewService(mockReviewRepo, mockCardRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	cardID := int64(100)

	t.Run("Success", func(t *testing.T) {
		c, _ := card.NewBuilder().
			WithID(cardID).
			WithState(valueobjects.CardStateNew).
			Build()

		mockTM.ExpectTransaction()
		mockCardRepo.On("FindByID", mock.Anything, userID, cardID).Return(c, nil).Once()
		mockCardRepo.On("Update", mock.Anything, userID, cardID, mock.Anything).Return(nil).Once()
		mockReviewRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*review.Review")).Return(nil).Once()

		result, err := service.Create(ctx, userID, cardID, 3, 5000) // Rating Good (3), 5 seconds

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, cardID, result.GetCardID())
		assert.Equal(t, 3, result.GetRating())
		
		// Verify card state changed from New to Learn
		assert.Equal(t, valueobjects.CardStateLearn, c.GetState())
		
		mockReviewRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Card Not Found", func(t *testing.T) {
		mockTM.ExpectTransaction()
		mockCardRepo.On("FindByID", mock.Anything, userID, cardID).Return(nil, nil).Once()

		result, err := service.Create(ctx, userID, cardID, 3, 5000)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "card not found")
		mockTM.AssertExpectations(t)
	})
}

