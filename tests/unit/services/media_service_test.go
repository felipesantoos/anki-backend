package services

import (
	"context"
	"testing"

	mediaSvc "github.com/felipesantos/anki-backend/core/services/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMediaService_Create(t *testing.T) {
	mockRepo := new(MockMediaRepository)
	service := mediaSvc.NewMediaService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		filename := "img.png"
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, filename, "hash123", 100, "image/png", "/path/to/storage")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, filename, result.GetFilename())
		mockRepo.AssertExpectations(t)
	})
}

