package services

import (
	"context"
	"testing"

	filteredSvc "github.com/felipesantos/anki-backend/core/services/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFilteredDeckService_Create(t *testing.T) {
	mockRepo := new(MockFilteredDeckRepository)
	service := filteredSvc.NewFilteredDeckService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "Filtered"
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, "tag:test", 20, "oldest", true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})
}

