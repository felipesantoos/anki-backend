package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
)

// IBackupService defines the interface for backup management
type IBackupService interface {
	// Create records a new backup
	Create(ctx context.Context, userID int64, filename string, size int64, storagePath string, backupType string) (*backup.Backup, error)

	// FindByUserID finds all backups for a user
	FindByUserID(ctx context.Context, userID int64) ([]*backup.Backup, error)

	// Delete removes a backup record
	Delete(ctx context.Context, userID int64, id int64) error
}

