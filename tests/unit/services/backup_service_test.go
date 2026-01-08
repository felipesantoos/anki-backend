package services

import (
	"context"
	"testing"

	backupSvc "github.com/felipesantos/anki-backend/core/services/backup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBackupService_Create(t *testing.T) {
	mockRepo := new(MockBackupRepository)
	mockExportSvc := new(MockExportService)
	mockStorageRepo := new(MockStorageRepository)
	service := backupSvc.NewBackupService(mockRepo, mockExportSvc, mockStorageRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		filename := "backup_2023.colpkg"
		mockRepo.On("Save", ctx, userID, mock.Anything).Return(nil).Once()

		result, err := service.Create(ctx, userID, filename, 5000, "/backups/1", "manual")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, filename, result.GetFilename())
		mockRepo.AssertExpectations(t)
	})
}

