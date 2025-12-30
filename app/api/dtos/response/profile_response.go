package response

import "time"

// ProfileResponse represents the response payload for a profile
type ProfileResponse struct {
	ID           int64      `json:"id" example:"1"`
	UserID       int64      `json:"user_id" example:"1"`
	Name         string     `json:"name" example:"Pessoal"`
	SyncEnabled  bool       `json:"sync_enabled"`
	SyncUsername *string    `json:"sync_username,omitempty" example:"user@ankiweb.net"`
	CreatedAt    time.Time  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt    time.Time  `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

