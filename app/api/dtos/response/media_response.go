package response

import "time"

// MediaResponse represents the response payload for a media file
type MediaResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Filename    string    `json:"filename"`
	Hash        string    `json:"hash"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	StoragePath string    `json:"storage_path"`
	CreatedAt   time.Time `json:"created_at"`
}

