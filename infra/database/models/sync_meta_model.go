package models

import (
	"time"
)

// SyncMetaModel represents the sync_meta table structure in the database
type SyncMetaModel struct {
	ID          int64
	UserID      int64
	ClientID    string
	LastSync    time.Time
	LastSyncUSN int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

