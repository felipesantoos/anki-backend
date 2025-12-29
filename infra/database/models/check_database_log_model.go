package models

import (
	"database/sql"
	"time"
)

// CheckDatabaseLogModel represents the check_database_log table structure in the database
type CheckDatabaseLogModel struct {
	ID              int64
	UserID          int64
	Status          string
	IssuesFound     int
	IssuesDetails   string // JSONB stored as string
	ExecutionTimeMs sql.NullInt64
	CreatedAt       time.Time
}

