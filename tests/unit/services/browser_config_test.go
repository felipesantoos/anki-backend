package services

import (
	"context"
	"testing"

	browserconfig "github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
	browserconfigsvc "github.com/felipesantos/anki-backend/core/services/browserconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBrowserConfigService_FindByUserID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockBrowserConfigRepository)
	service := browserconfigsvc.NewBrowserConfigService(mockRepo)

	userID := int64(1)

	t.Run("Success_ExistingConfig", func(t *testing.T) {
		existingCfg, _ := browserconfig.NewBuilder().
			WithUserID(userID).
			WithVisibleColumns([]string{"note"}).
			Build()

		mockRepo.On("FindByUserID", ctx, userID).Return(existingCfg, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, existingCfg, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success_CreateDefaults", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.AnythingOfType("*browserconfig.BrowserConfig")).Return(nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.GetUserID())
		assert.Contains(t, result.GetVisibleColumns(), "note")
		mockRepo.AssertExpectations(t)
	})
}

func TestBrowserConfigService_Update(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockBrowserConfigRepository)
	service := browserconfigsvc.NewBrowserConfigService(mockRepo)

	userID := int64(1)
	cfgID := int64(100)
	cfg, _ := browserconfig.NewBuilder().
		WithUserID(userID).
		WithVisibleColumns([]string{"note"}).
		Build()
	cfg.SetID(cfgID)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("Update", ctx, userID, cfgID, cfg).Return(nil).Once()

		err := service.Update(ctx, userID, cfg)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

