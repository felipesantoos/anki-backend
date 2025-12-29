package models

import (
	"time"
)

// DeletionLogModel represents the deletions_log table structure in the database
type DeletionLogModel struct {
	ID         int64
	UserID     int64
	ObjectType string
	ObjectID   int64
	ObjectData string // JSONB stored as string
	DeletedAt  time.Time
}

