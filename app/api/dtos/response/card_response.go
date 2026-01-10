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

// ListCardsResponse represents the response payload for listing cards
type ListCardsResponse struct {
	// List of cards
	Data []*CardResponse `json:"data"`

	// Pagination metadata
	Pagination PaginationResponse `json:"pagination"`
}

// CardInfoResponse represents detailed card information (Card Info Dialog)
// @Description Detailed card information including note data, deck/note type names, and review history
type CardInfoResponse struct {
	// Card ID
	CardID int64 `json:"card_id" example:"1"`

	// Note ID
	NoteID int64 `json:"note_id" example:"1"`

	// Deck name
	DeckName string `json:"deck_name" example:"Default"`

	// Note type name
	NoteTypeName string `json:"note_type_name" example:"Basic"`

	// Note fields (parsed from fieldsJSON)
	Fields map[string]interface{} `json:"fields" example:"{\"Front\":\"Hello\",\"Back\":\"Olá\"}"`

	// Note tags
	Tags []string `json:"tags" example:"[\"vocabulary\"]"`

	// Card creation timestamp
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// First review timestamp (null if no reviews)
	FirstReview *time.Time `json:"first_review,omitempty" example:"2024-01-16T10:00:00Z"`

	// Last review timestamp (null if no reviews)
	LastReview *time.Time `json:"last_review,omitempty" example:"2024-01-20T15:45:00Z"`

	// Total number of reviews
	TotalReviews int `json:"total_reviews" example:"5"`

	// Ease factor history (array of ease values ordered by review date)
	EaseHistory []int `json:"ease_history" example:"[2500,2600,2500,2400,2500]"`

	// Interval history (array of interval values ordered by review date)
	IntervalHistory []int `json:"interval_history" example:"[1,2,4,1,2]"`

	// Full review history
	ReviewHistory []CardInfoReviewItem `json:"review_history"`
}

// CardInfoReviewItem represents a single review in the card info history
type CardInfoReviewItem struct {
	// Rating given (1-4)
	Rating int `json:"rating" example:"3"`

	// Interval after review (days or negative seconds)
	Interval int `json:"interval" example:"1"`

	// Ease factor after review (permille)
	Ease int `json:"ease" example:"2500"`

	// Time spent on review (milliseconds)
	TimeMs int `json:"time_ms" example:"5000"`

	// Review type (new, learn, review, relearn)
	Type string `json:"type" example:"review"`

	// Review timestamp
	CreatedAt time.Time `json:"created_at" example:"2024-01-20T15:45:00Z"`
}

// CardPositionResponse represents the response payload for a card's position
type CardPositionResponse struct {
	// Ordinal position
	Position int `json:"position" example:"100"`
}
