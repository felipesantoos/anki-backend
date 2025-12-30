package services

import (
	"context"
	"testing"

	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	syncSvc "github.com/felipesantos/anki-backend/core/services/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncMetaService_FindByUserID(t *testing.T) {
	mockRepo := new(MockSyncMetaRepository)
	service := syncSvc.NewSyncMetaService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Existing Meta", func(t *testing.T) {
		meta, _ := syncmeta.NewBuilder().WithUserID(userID).WithLastSyncUSN(10).Build()
		mockRepo.On("FindByUserID", ctx, userID).Return([]*syncmeta.SyncMeta{meta}, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, int64(10), result.GetLastSyncUSN())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return([]*syncmeta.SyncMeta{}, nil).Once()

		result, err := service.FindByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestSyncMetaService_Update(t *testing.T) {
	mockRepo := new(MockSyncMetaRepository)
	service := syncSvc.NewSyncMetaService(mockRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Create New", func(t *testing.T) {
		mockRepo.On("FindByUserID", ctx, userID).Return([]*syncmeta.SyncMeta{}, nil).Once()
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Update(ctx, userID, "client1", 5)

		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.GetLastSyncUSN())
		assert.Equal(t, "client1", result.GetClientID())
		mockRepo.AssertExpectations(t)
	})
}
