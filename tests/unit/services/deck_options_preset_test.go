package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
	presetSvc "github.com/felipesantos/anki-backend/core/services/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeckOptionsPresetService_Create(t *testing.T) {
	mockRepo := new(MockDeckOptionsPresetRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo, mockDeckRepo, mockTM)
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
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo, mockDeckRepo, mockTM)
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
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo, mockDeckRepo, mockTM)
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
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo, mockDeckRepo, mockTM)
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

func TestDeckOptionsPresetService_ApplyToDecks(t *testing.T) {
	mockRepo := new(MockDeckOptionsPresetRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := presetSvc.NewDeckOptionsPresetService(mockRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	presetID := int64(10)
	deckIDs := []int64{100, 101}

	preset, _ := deckoptionspreset.NewBuilder().WithID(presetID).WithUserID(userID).WithName("Preset").WithOptionsJSON(`{"key":"value"}`).Build()
	deck1, _ := deck.NewBuilder().WithID(100).WithUserID(userID).WithName("Deck 1").Build()
	deck2, _ := deck.NewBuilder().WithID(101).WithUserID(userID).WithName("Deck 2").Build()

	t.Run("Success", func(t *testing.T) {
		mockTM.ExpectTransaction()
		mockRepo.On("FindByID", ctx, userID, presetID).Return(preset, nil).Once()
		mockDeckRepo.On("FindByID", ctx, userID, int64(100)).Return(deck1, nil).Once()
		mockDeckRepo.On("Update", ctx, userID, int64(100), mock.Anything).Return(nil).Once()
		mockDeckRepo.On("FindByID", ctx, userID, int64(101)).Return(deck2, nil).Once()
		mockDeckRepo.On("Update", ctx, userID, int64(101), mock.Anything).Return(nil).Once()

		err := service.ApplyToDecks(ctx, userID, presetID, deckIDs)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Preset Not Found", func(t *testing.T) {
		mockTM.ExpectTransaction()
		mockRepo.On("FindByID", ctx, userID, presetID).Return(nil, nil).Once()

		err := service.ApplyToDecks(ctx, userID, presetID, deckIDs)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "preset not found")
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Deck Not Found", func(t *testing.T) {
		mockTM.ExpectTransaction()
		mockRepo.On("FindByID", ctx, userID, presetID).Return(preset, nil).Once()
		mockDeckRepo.On("FindByID", ctx, userID, int64(100)).Return(nil, nil).Once()

		err := service.ApplyToDecks(ctx, userID, presetID, deckIDs)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deck 100 not found")
		mockRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

