package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
)

// IBackupRepository defines the interface for backup data persistence
// All methods that access specific resources require userID to ensure data isolation
type IBackupRepository interface {
	// Save saves or updates a backup in the database
	// If the backup has an ID, it updates the existing backup
	// If the backup has no ID, it creates a new backup and returns it with the ID set
	Save(ctx context.Context, userID int64, backupEntity *backup.Backup) error

	// FindByID finds a backup by ID, filtering by userID to ensure ownership
	// Returns the backup if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*backup.Backup, error)

	// FindByUserID finds all backups for a user
	// Returns a list of backups belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*backup.Backup, error)

	// Update updates an existing backup, validating ownership
	// Returns error if backup doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, backupEntity *backup.Backup) error

	// Delete deletes a backup, validating ownership (hard delete - backups don't have soft delete)
	// Returns error if backup doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a backup exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByFilename finds a backup by filename, filtering by userID to ensure ownership
	FindByFilename(ctx context.Context, userID int64, filename string) (*backup.Backup, error)

	// FindByType finds all backups of a specific type for a user
	FindByType(ctx context.Context, userID int64, backupType string) ([]*backup.Backup, error)
}

