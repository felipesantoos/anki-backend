package models

import (
	"time"
)

// BackupModel represents the backups table structure in the database
type BackupModel struct {
	ID          int64
	UserID      int64
	Filename    string
	Size        int64
	StoragePath string
	BackupType  string
	CreatedAt   time.Time
}

