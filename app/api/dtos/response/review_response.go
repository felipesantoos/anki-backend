package response

import "time"

// ReviewResponse represents the response payload for a card review
// @Description Response payload containing review information
type ReviewResponse struct {
	// Unique identifier for the review
	ID int64 `json:"id" example:"1"`

	// ID of the card reviewed
	CardID int64 `json:"card_id" example:"1"`

	// Rating given (1-4)
	Rating int `json:"rating" example:"3"`

	// Time taken in milliseconds
	TimeMs int `json:"time_ms" example:"5000"`

	// Timestamp when review was performed
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

