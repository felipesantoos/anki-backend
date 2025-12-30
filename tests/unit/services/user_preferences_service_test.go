package services

import (
	"context"
	"testing"

	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	userPrefsSvc "github.com/felipesantos/anki-backend/core/services/userpreferences"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserPreferencesService_FindByUserID(t *testing.T) {
	mockRepo := new(MockUserPreferencesRepository)
	service := userPrefsSvc.NewUserPreferencesService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Existing Preferences", func(t *testing.T) {
		prefs, _ := userpreferences.NewBuilder().WithUserID(userID).WithLanguage("en").Build()
		mockRepo.On("FindByUserID", ctx, userID).Return(prefs, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, "en", result.GetLanguage())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Default Preferences Creation", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pt-BR", result.GetLanguage()) // Default from service
		mockRepo.AssertExpectations(t)
	})
}

