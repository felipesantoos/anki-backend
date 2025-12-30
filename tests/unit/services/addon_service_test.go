package services

import (
	"context"
	"testing"

	addonSvc "github.com/felipesantos/anki-backend/core/services/addon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddOnService_Create(t *testing.T) {
	mockRepo := new(MockAddOnRepository)
	service := addonSvc.NewAddOnService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		code := "12345"
		name := "Awesome Addon"
		mockRepo.On("FindByCode", ctx, userID, code).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Install(ctx, userID, code, name, "1.0", "{}")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})
}

