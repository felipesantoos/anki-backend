package models

import (
	"time"
)

// UndoHistoryModel represents the undo_history table structure in the database
type UndoHistoryModel struct {
	ID            int64
	UserID        int64
	OperationType string
	OperationData string // JSONB stored as string
	CreatedAt     time.Time
}

