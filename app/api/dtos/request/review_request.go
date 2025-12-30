package request

// CreateReviewRequest represents the request payload to record a card review
// @Description Request payload for recording a card review
type CreateReviewRequest struct {
	// ID of the card being reviewed
	CardID int64 `json:"card_id" example:"1" validate:"required"`

	// Rating given to the card (1: Again, 2: Hard, 3: Good, 4: Easy)
	Rating int `json:"rating" example:"3" validate:"required,min=1,max=4"`

	// Time taken to answer in milliseconds
	TimeMs int `json:"time_ms" example:"5000" validate:"required,min=0"`
}

