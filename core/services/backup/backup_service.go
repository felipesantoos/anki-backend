package backup

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// BackupService implements IBackupService
type BackupService struct {
	repo secondary.IBackupRepository
}

// NewBackupService creates a new BackupService instance
func NewBackupService(repo secondary.IBackupRepository) primary.IBackupService {
	return &BackupService{
		repo: repo,
	}
}

// Create records a new backup
func (s *BackupService) Create(ctx context.Context, userID int64, filename string, size int64, storagePath string, backupType string) (*backup.Backup, error) {
	now := time.Now()
	b, err := backup.NewBuilder().
		WithUserID(userID).
		WithFilename(filename).
		WithSize(size).
		WithStoragePath(storagePath).
		WithBackupType(backupType).
		WithCreatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, b); err != nil {
		return nil, err
	}

	return b, nil
}

// FindByUserID finds all backups for a user
func (s *BackupService) FindByUserID(ctx context.Context, userID int64) ([]*backup.Backup, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// Delete removes a backup record
func (s *BackupService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

