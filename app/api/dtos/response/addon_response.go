package response

import "time"

// AddOnResponse represents the response payload for an add-on
type AddOnResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Version    string    `json:"version"`
	ConfigJSON string    `json:"config_json"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

