package mappers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestBackupToDomain_WithAllFields(t *testing.T) {
	now := time.Now()

	model := &models.BackupModel{
		ID:          1,
		UserID:      100,
		Filename:    "backup_20240101.apkg",
		Size:        5000000,
		StoragePath: "/storage/backups/backup_20240101.apkg",
		BackupType:  "automatic",
		CreatedAt:   now,
	}

	entity, err := mappers.BackupToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "backup_20240101.apkg", entity.GetFilename())
	assert.Equal(t, int64(5000000), entity.GetSize())
	assert.Equal(t, "/storage/backups/backup_20240101.apkg", entity.GetStoragePath())
	assert.Equal(t, "automatic", entity.GetBackupType())
	assert.Equal(t, now, entity.GetCreatedAt())
}

func TestBackupToDomain_AllBackupTypes(t *testing.T) {
	backupTypes := []string{
		"automatic",
		"manual",
		"pre_operation",
	}

	for _, backupType := range backupTypes {
		t.Run(backupType, func(t *testing.T) {
			model := &models.BackupModel{
				ID:          1,
				UserID:      100,
				Filename:    "backup.apkg",
				Size:        1000000,
				StoragePath: "/storage/backup.apkg",
				BackupType:  backupType,
				CreatedAt:   time.Now(),
			}

			entity, err := mappers.BackupToDomain(model)
			require.NoError(t, err)
			assert.Equal(t, backupType, entity.GetBackupType())
		})
	}
}

func TestBackupToDomain_NilInput(t *testing.T) {
	entity, err := BackupToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestBackupToModel_WithAllFields(t *testing.T) {
	now := time.Now()

	entity, err := backup.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithFilename("backup_20240101.apkg").
		WithSize(5000000).
		WithStoragePath("/storage/backups/backup_20240101.apkg").
		WithBackupType("automatic").
		WithCreatedAt(now).
		Build()
	require.NoError(t, err)

	model := mappers.BackupToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "backup_20240101.apkg", model.Filename)
	assert.Equal(t, int64(5000000), model.Size)
	assert.Equal(t, "/storage/backups/backup_20240101.apkg", model.StoragePath)
	assert.Equal(t, "automatic", model.BackupType)
	assert.Equal(t, now, model.CreatedAt)
}

