package response

import "time"

// DeckOptionsPresetResponse represents the response payload for a deck options preset
// @Description Response payload containing deck options preset information
type DeckOptionsPresetResponse struct {
	// Unique identifier for the preset
	ID int64 `json:"id" example:"1"`

	// ID of the user who owns the preset
	UserID int64 `json:"user_id" example:"1"`

	// Name of the preset
	Name string `json:"name" example:"Default Study"`

	// Configuration options as a JSON string
	OptionsJSON string `json:"options_json" example:"{}"`

	// Timestamp when the preset was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Timestamp when the preset was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

