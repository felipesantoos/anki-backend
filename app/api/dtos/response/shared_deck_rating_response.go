package response

import "time"

// SharedDeckRatingResponse represents the response payload for a shared deck rating
type SharedDeckRatingResponse struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	SharedDeckID int64     `json:"shared_deck_id"`
	Rating       int       `json:"rating"`
	Comment      *string   `json:"comment,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

