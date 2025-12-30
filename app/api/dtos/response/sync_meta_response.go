package response

import "time"

// SyncMetaResponse represents the response payload for sync metadata
type SyncMetaResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	ClientID    string    `json:"client_id"`
	LastSync    time.Time `json:"last_sync"`
	LastSyncUSN int64     `json:"last_sync_usn"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

