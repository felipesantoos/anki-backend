package services

import (
	"context"
	"testing"

	presetSvc "github.com/felipesantos/anki-backend/core/services/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeckOptionsPresetService_Create(t *testing.T) {
	mockRepo := new(MockDeckOptionsPresetRepository)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "My Preset"
		mockRepo.On("FindByName", ctx, userID, name).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, name, "{}")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})
}

