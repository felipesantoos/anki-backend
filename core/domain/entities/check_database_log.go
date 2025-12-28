package entities

import (
	"time"
)

// CheckStatus represents the status of a database check
const (
	CheckStatusRunning   = "running"
	CheckStatusCompleted = "completed"
	CheckStatusFailed    = "failed"
	CheckStatusCorrupted = "corrupted"
)

// CheckDatabaseLog represents a database check log entry entity in the domain
// It stores information about database integrity checks
type CheckDatabaseLog struct {
	ID              int64
	UserID          int64
	Status          string // running, completed, failed, corrupted
	IssuesFound     int
	IssuesDetails   string // JSONB in database
	ExecutionTimeMs *int
	CreatedAt       time.Time
}

// IsCompleted checks if the check is completed
func (cdl *CheckDatabaseLog) IsCompleted() bool {
	return cdl.Status == CheckStatusCompleted
}

// IsFailed checks if the check failed
func (cdl *CheckDatabaseLog) IsFailed() bool {
	return cdl.Status == CheckStatusFailed
}

// IsCorrupted checks if the database is corrupted
func (cdl *CheckDatabaseLog) IsCorrupted() bool {
	return cdl.Status == CheckStatusCorrupted
}

