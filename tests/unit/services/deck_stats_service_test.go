package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	deckSvc "github.com/felipesantos/anki-backend/core/services/deck"
	"github.com/felipesantos/anki-backend/pkg/ownership"
	"github.com/stretchr/testify/assert"
)

func TestDeckStatsService_GetStats(t *testing.T) {
	mockRepo := new(MockDeckRepository)
	service := deckSvc.NewDeckStatsService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	deckID := int64(10)

	expectedStats := &deck.DeckStats{
		DeckID:         deckID,
		NewCount:       5,
		LearningCount:  2,
		ReviewCount:    10,
		SuspendedCount: 1,
		NotesCount:     15,
		DueTodayCount:  3,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetStats", ctx, userID, deckID).Return(expectedStats, nil).Once()

		result, err := service.GetStats(ctx, userID, deckID)

		assert.NoError(t, err)
		assert.Equal(t, expectedStats, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("GetStats", ctx, userID, deckID).Return(nil, ownership.ErrResourceNotFound).Once()

		result, err := service.GetStats(ctx, userID, deckID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

