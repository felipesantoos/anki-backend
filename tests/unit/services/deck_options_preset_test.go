package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
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

func TestDeckOptionsPresetService_FindByUserID(t *testing.T) {
	mockRepo := new(MockDeckOptionsPresetRepository)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return([]*deckoptionspreset.DeckOptionsPreset{}, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckOptionsPresetService_Update(t *testing.T) {
	mockRepo := new(MockDeckOptionsPresetRepository)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	presetID := int64(10)

	existingPreset, _ := deckoptionspreset.NewBuilder().WithID(presetID).WithUserID(userID).WithName("Old Name").Build()

	t.Run("Success", func(t *testing.T) {
		newName := "New Name"
		mockRepo.On("FindByID", ctx, userID, presetID).Return(existingPreset, nil).Once()
		mockRepo.On("FindByName", ctx, userID, newName).Return(nil, nil).Once()
		mockRepo.On("Update", ctx, userID, presetID, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, userID, presetID, newName, "{}")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.GetName())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Name Conflict", func(t *testing.T) {
		newName := "Conflict Name"
		conflictPreset, _ := deckoptionspreset.NewBuilder().WithID(20).WithUserID(userID).WithName(newName).Build()

		mockRepo.On("FindByID", ctx, userID, presetID).Return(existingPreset, nil).Once()
		mockRepo.On("FindByName", ctx, userID, newName).Return(conflictPreset, nil).Once()

		result, err := service.Update(ctx, userID, presetID, newName, "{}")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, presetID).Return(nil, nil).Once()

		result, err := service.Update(ctx, userID, presetID, "Some Name", "{}")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "preset not found")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeckOptionsPresetService_Delete(t *testing.T) {
	mockRepo := new(MockDeckOptionsPresetRepository)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo)
	ctx := context.Background()
	userID := int64(1)
	presetID := int64(10)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Delete", ctx, userID, presetID).Return(nil).Once()

		err := service.Delete(ctx, userID, presetID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

