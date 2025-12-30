package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	profileSvc "github.com/felipesantos/anki-backend/core/services/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProfileService_Create(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service := profileSvc.NewProfileService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		name := "Work"
		mockRepo.On("FindByName", ctx, userID, name).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*profile.Profile")).Return(nil).Once()

		result, err := service.Create(ctx, userID, name)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.GetName())
		mockRepo.AssertExpectations(t)
	})
}

func TestProfileService_EnableSync(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service := profileSvc.NewProfileService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	profileID := int64(100)

	t.Run("Success", func(t *testing.T) {
		p, _ := profile.NewBuilder().WithID(profileID).WithUserID(userID).WithName("Main").Build()
		username := "ankiuser"

		mockRepo.On("FindByID", ctx, userID, profileID).Return(p, nil).Once()
		mockRepo.On("Update", ctx, userID, profileID, mock.Anything).Return(nil).Once()

		err := service.EnableSync(ctx, userID, profileID, username)

		assert.NoError(t, err)
		assert.True(t, p.GetAnkiWebSyncEnabled())
		assert.Equal(t, &username, p.GetAnkiWebUsername())
		mockRepo.AssertExpectations(t)
	})
}

