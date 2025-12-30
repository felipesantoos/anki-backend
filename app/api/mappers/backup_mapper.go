package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
)

// ToBackupResponse converts Backup domain entity to Response DTO
func ToBackupResponse(b *backup.Backup) *response.BackupResponse {
	if b == nil {
		return nil
	}
	return &response.BackupResponse{
		ID:          b.GetID(),
		UserID:      b.GetUserID(),
		Filename:    b.GetFilename(),
		Size:        b.GetSize(),
		StoragePath: b.GetStoragePath(),
		BackupType:  b.GetBackupType(),
		CreatedAt:   b.GetCreatedAt(),
	}
}

// ToBackupResponseList converts list of Backup domain entities to list of Response DTOs
func ToBackupResponseList(backups []*backup.Backup) []*response.BackupResponse {
	res := make([]*response.BackupResponse, len(backups))
	for i, b := range backups {
		res[i] = ToBackupResponse(b)
	}
	return res
}

