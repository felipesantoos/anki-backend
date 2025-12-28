package entities

import (
	"time"
)

// BackupType represents the type of backup
const (
	BackupTypeAutomatic    = "automatic"
	BackupTypeManual       = "manual"
	BackupTypePreOperation = "pre_operation"
)

// Backup represents a backup entity in the domain
// It stores information about user backups
type Backup struct {
	ID          int64
	UserID      int64
	Filename    string
	Size        int64
	StoragePath string
	BackupType  string // automatic, manual, pre_operation
	CreatedAt   time.Time
}

// IsAutomatic checks if the backup is automatic
func (b *Backup) IsAutomatic() bool {
	return b.BackupType == BackupTypeAutomatic
}

// IsManual checks if the backup is manual
func (b *Backup) IsManual() bool {
	return b.BackupType == BackupTypeManual
}

// IsPreOperation checks if the backup is a pre-operation backup
func (b *Backup) IsPreOperation() bool {
	return b.BackupType == BackupTypePreOperation
}

