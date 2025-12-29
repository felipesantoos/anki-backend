package mappers

import (
	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// BackupToDomain converts a BackupModel (database representation) to a Backup entity (domain representation)
func BackupToDomain(model *models.BackupModel) (*backup.Backup, error) {
	if model == nil {
		return nil, nil
	}

	builder := backup.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithFilename(model.Filename).
		WithSize(model.Size).
		WithStoragePath(model.StoragePath).
		WithBackupType(model.BackupType).
		WithCreatedAt(model.CreatedAt)

	return builder.Build()
}

// BackupToModel converts a Backup entity (domain representation) to a BackupModel (database representation)
func BackupToModel(backupEntity *backup.Backup) *models.BackupModel {
	return &models.BackupModel{
		ID:          backupEntity.GetID(),
		UserID:      backupEntity.GetUserID(),
		Filename:    backupEntity.GetFilename(),
		Size:        backupEntity.GetSize(),
		StoragePath: backupEntity.GetStoragePath(),
		BackupType:  backupEntity.GetBackupType(),
		CreatedAt:   backupEntity.GetCreatedAt(),
	}
}

