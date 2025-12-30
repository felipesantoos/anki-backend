package response

import "time"

// BackupResponse represents the response payload for a backup
type BackupResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	StoragePath string    `json:"storage_path"`
	BackupType  string    `json:"backup_type"`
	CreatedAt   time.Time `json:"created_at"`
}

