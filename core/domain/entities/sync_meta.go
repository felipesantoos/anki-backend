package entities

import (
	"time"
)

// SyncMeta represents synchronization metadata entity in the domain
// It tracks sync state for each user and client device
type SyncMeta struct {
	ID          int64
	UserID      int64
	ClientID    string
	LastSync    time.Time
	LastSyncUSN int64 // Update Sequence Number
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

