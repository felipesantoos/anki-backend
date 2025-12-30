package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// BackupRepository implements IBackupRepository using PostgreSQL
type BackupRepository struct {
	db *sql.DB
}

// NewBackupRepository creates a new BackupRepository instance
func NewBackupRepository(db *sql.DB) secondary.IBackupRepository {
	return &BackupRepository{
		db: db,
	}
}

// Save saves or updates a backup in the database
func (r *BackupRepository) Save(ctx context.Context, userID int64, backupEntity *backup.Backup) error {
	model := mappers.BackupToModel(backupEntity)

	if backupEntity.GetID() == 0 {
		// Insert new backup
		query := `
			INSERT INTO backups (user_id, filename, size, storage_path, backup_type, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}

		var backupID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Filename,
			model.Size,
			model.StoragePath,
			model.BackupType,
			model.CreatedAt,
		).Scan(&backupID)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		backupEntity.SetID(backupID)
		return nil
	}

	// Update existing backup - validate ownership first
	existingBackup, err := r.FindByID(ctx, userID, backupEntity.GetID())
	if err != nil {
		return err
	}
	if existingBackup == nil {
		return ownership.ErrResourceNotFound
	}

	// Update backup
	query := `
		UPDATE backups
		SET filename = $1, size = $2, storage_path = $3, backup_type = $4
		WHERE id = $5 AND user_id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		model.Filename,
		model.Size,
		model.StoragePath,
		model.BackupType,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update backup: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// FindByID finds a backup by ID, filtering by userID to ensure ownership
func (r *BackupRepository) FindByID(ctx context.Context, userID int64, id int64) (*backup.Backup, error) {
	query := `
		SELECT id, user_id, filename, size, storage_path, backup_type, created_at
		FROM backups
		WHERE id = $1 AND user_id = $2
	`

	var model models.BackupModel
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Filename,
		&model.Size,
		&model.StoragePath,
		&model.BackupType,
		&model.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find backup: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.BackupToDomain(&model)
}

// FindByUserID finds all backups for a user
func (r *BackupRepository) FindByUserID(ctx context.Context, userID int64) ([]*backup.Backup, error) {
	query := `
		SELECT id, user_id, filename, size, storage_path, backup_type, created_at
		FROM backups
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find backups by user ID: %w", err)
	}
	defer rows.Close()

	var backups []*backup.Backup
	for rows.Next() {
		var model models.BackupModel

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Filename,
			&model.Size,
			&model.StoragePath,
			&model.BackupType,
			&model.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backup: %w", err)
		}

		backupEntity, err := mappers.BackupToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert backup to domain: %w", err)
		}
		backups = append(backups, backupEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating backups: %w", err)
	}

	return backups, nil
}

// Update updates an existing backup, validating ownership
func (r *BackupRepository) Update(ctx context.Context, userID int64, id int64, backupEntity *backup.Backup) error {
	return r.Save(ctx, userID, backupEntity)
}

// Delete deletes a backup, validating ownership
func (r *BackupRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingBackup, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingBackup == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (backups don't have soft delete)
	query := `DELETE FROM backups WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// Exists checks if a backup exists and belongs to the user
func (r *BackupRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM backups
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check backup existence: %w", err)
	}

	return exists, nil
}

// FindByFilename finds a backup by filename, filtering by userID to ensure ownership
func (r *BackupRepository) FindByFilename(ctx context.Context, userID int64, filename string) (*backup.Backup, error) {
	query := `
		SELECT id, user_id, filename, size, storage_path, backup_type, created_at
		FROM backups
		WHERE filename = $1 AND user_id = $2
	`

	var model models.BackupModel
	err := r.db.QueryRowContext(ctx, query, filename, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Filename,
		&model.Size,
		&model.StoragePath,
		&model.BackupType,
		&model.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find backup by filename: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.BackupToDomain(&model)
}

// FindByType finds all backups of a specific type for a user
func (r *BackupRepository) FindByType(ctx context.Context, userID int64, backupType string) ([]*backup.Backup, error) {
	query := `
		SELECT id, user_id, filename, size, storage_path, backup_type, created_at
		FROM backups
		WHERE user_id = $1 AND backup_type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, backupType)
	if err != nil {
		return nil, fmt.Errorf("failed to find backups by type: %w", err)
	}
	defer rows.Close()

	var backups []*backup.Backup
	for rows.Next() {
		var model models.BackupModel
		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Filename,
			&model.Size,
			&model.StoragePath,
			&model.BackupType,
			&model.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backup: %w", err)
		}

		backupEntity, err := mappers.BackupToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert backup to domain: %w", err)
		}
		backups = append(backups, backupEntity)
	}

	return backups, nil
}

// Ensure BackupRepository implements IBackupRepository
var _ secondary.IBackupRepository = (*BackupRepository)(nil)

