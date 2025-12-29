package backup

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
	id          int64
	userID      int64
	filename    string
	size        int64
	storagePath string
	backupType  string // automatic, manual, pre_operation
	createdAt   time.Time
}

// Getters
func (b *Backup) GetID() int64 {
	return b.id
}

func (b *Backup) GetUserID() int64 {
	return b.userID
}

func (b *Backup) GetFilename() string {
	return b.filename
}

func (b *Backup) GetSize() int64 {
	return b.size
}

func (b *Backup) GetStoragePath() string {
	return b.storagePath
}

func (b *Backup) GetBackupType() string {
	return b.backupType
}

func (b *Backup) GetCreatedAt() time.Time {
	return b.createdAt
}

// Setters
func (b *Backup) SetID(id int64) {
	b.id = id
}

func (b *Backup) SetUserID(userID int64) {
	b.userID = userID
}

func (b *Backup) SetFilename(filename string) {
	b.filename = filename
}

func (b *Backup) SetSize(size int64) {
	b.size = size
}

func (b *Backup) SetStoragePath(storagePath string) {
	b.storagePath = storagePath
}

func (b *Backup) SetBackupType(backupType string) {
	b.backupType = backupType
}

func (b *Backup) SetCreatedAt(createdAt time.Time) {
	b.createdAt = createdAt
}

// IsAutomatic checks if the backup is automatic
func (b *Backup) IsAutomatic() bool {
	return b.backupType == BackupTypeAutomatic
}

// IsManual checks if the backup is manual
func (b *Backup) IsManual() bool {
	return b.backupType == BackupTypeManual
}

// IsPreOperation checks if the backup is a pre-operation backup
func (b *Backup) IsPreOperation() bool {
	return b.backupType == BackupTypePreOperation
}

