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
	id              int64
	userID          int64
	status          string // running, completed, failed, corrupted
	issuesFound     int
	issuesDetails   string // JSONB in database
	executionTimeMs *int
	createdAt       time.Time
}

// Getters
func (cdl *CheckDatabaseLog) GetID() int64 {
	return cdl.id
}

func (cdl *CheckDatabaseLog) GetUserID() int64 {
	return cdl.userID
}

func (cdl *CheckDatabaseLog) GetStatus() string {
	return cdl.status
}

func (cdl *CheckDatabaseLog) GetIssuesFound() int {
	return cdl.issuesFound
}

func (cdl *CheckDatabaseLog) GetIssuesDetails() string {
	return cdl.issuesDetails
}

func (cdl *CheckDatabaseLog) GetExecutionTimeMs() *int {
	return cdl.executionTimeMs
}

func (cdl *CheckDatabaseLog) GetCreatedAt() time.Time {
	return cdl.createdAt
}

// Setters
func (cdl *CheckDatabaseLog) SetID(id int64) {
	cdl.id = id
}

func (cdl *CheckDatabaseLog) SetUserID(userID int64) {
	cdl.userID = userID
}

func (cdl *CheckDatabaseLog) SetStatus(status string) {
	cdl.status = status
}

func (cdl *CheckDatabaseLog) SetIssuesFound(issuesFound int) {
	cdl.issuesFound = issuesFound
}

func (cdl *CheckDatabaseLog) SetIssuesDetails(issuesDetails string) {
	cdl.issuesDetails = issuesDetails
}

func (cdl *CheckDatabaseLog) SetExecutionTimeMs(executionTimeMs *int) {
	cdl.executionTimeMs = executionTimeMs
}

func (cdl *CheckDatabaseLog) SetCreatedAt(createdAt time.Time) {
	cdl.createdAt = createdAt
}

// IsCompleted checks if the check is completed
func (cdl *CheckDatabaseLog) IsCompleted() bool {
	return cdl.status == CheckStatusCompleted
}

// IsFailed checks if the check failed
func (cdl *CheckDatabaseLog) IsFailed() bool {
	return cdl.status == CheckStatusFailed
}

// IsCorrupted checks if the database is corrupted
func (cdl *CheckDatabaseLog) IsCorrupted() bool {
	return cdl.status == CheckStatusCorrupted
}

