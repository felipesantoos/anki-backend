package response

import "time"

// DeckResponse represents the response payload for a deck
// @Description Response payload containing deck information
type DeckResponse struct {
	// Unique identifier for the deck
	ID int64 `json:"id" example:"1"`

	// ID of the user who owns the deck
	UserID int64 `json:"user_id" example:"1"`

	// Name of the deck
	Name string `json:"name" example:"Idiomas::Inglês"`

	// ID of the parent deck (null if it's a root deck)
	ParentID *int64 `json:"parent_id"`

	// Configuration options as a JSON string
	OptionsJSON string `json:"options_json" example:"{}"`

	// Timestamp when the deck was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Timestamp when the deck was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

