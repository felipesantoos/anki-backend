package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// BackupService implements IBackupService
type BackupService struct {
	repo          secondary.IBackupRepository
	exportSvc     primary.IExportService
	storageRepo   secondary.IStorageRepository
}

// NewBackupService creates a new BackupService instance
func NewBackupService(
	repo secondary.IBackupRepository,
	exportSvc primary.IExportService,
	storageRepo secondary.IStorageRepository,
) primary.IBackupService {
	return &BackupService{
		repo:        repo,
		exportSvc:   exportSvc,
		storageRepo: storageRepo,
	}
}

// Create records a new backup manually
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

// CreatePreOperationBackup creates an automatic backup before a destructive operation
func (s *BackupService) CreatePreOperationBackup(ctx context.Context, userID int64) (*backup.Backup, error) {
	// 1. Export user data
	reader, size, err := s.exportSvc.ExportCollection(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to export collection: %w", err)
	}

	// 2. Upload to storage
	filename := fmt.Sprintf("pre_op_%d_%d.json", userID, time.Now().Unix())
	storagePath := fmt.Sprintf("backups/%d/%s", userID, filename)
	
	fileInfo, err := s.storageRepo.Upload(ctx, reader, storagePath, "application/json")
	if err != nil {
		return nil, fmt.Errorf("failed to upload backup to storage: %w", err)
	}

	// 3. Save metadata to DB
	return s.Create(ctx, userID, filename, size, fileInfo.Path, backup.BackupTypePreOperation)
}

// FindByUserID finds all backups for a user
func (s *BackupService) FindByUserID(ctx context.Context, userID int64) ([]*backup.Backup, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// Delete removes a backup record
func (s *BackupService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

