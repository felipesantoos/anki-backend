package response

import "time"

// CardResponse represents the response payload for a card
// @Description Response payload containing card information
type CardResponse struct {
	// Unique identifier for the card
	ID int64 `json:"id" example:"1"`

	// ID of the note this card belongs to
	NoteID int64 `json:"note_id" example:"1"`

	// ID of the deck this card belongs to
	DeckID int64 `json:"deck_id" example:"1"`

	// Current state (new, learn, review, relearn)
	State string `json:"state" example:"new"`

	// Interval until next review (in days)
	Interval int `json:"interval" example:"1"`

	// Ease factor (in per mille)
	Ease int `json:"ease" example:"2500"`

	// Number of reviews
	Reviews int `json:"reviews" example:"0"`

	// Number of lapses
	Lapses int `json:"lapses" example:"0"`

	// Next review date (Unix timestamp)
	Due int64 `json:"due" example:"1705324200"`

	// Ordinal position
	Ord int `json:"ord" example:"0"`

	// Colored flag (0: none)
	Flags int `json:"flags" example:"0"`

	// Whether suspended
	Suspended bool `json:"suspended" example:"false"`

	// Timestamp when created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Timestamp when last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

