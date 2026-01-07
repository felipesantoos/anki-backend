package request

// UpdateCardRequest represents the request payload to update an existing card
// @Description Request payload for updating an existing card
type UpdateCardRequest struct {
	// ID of the deck the card belongs to
	DeckID int64 `json:"deck_id" example:"1" validate:"required"`

	// Current state of the card (e.g., 0: new, 1: learning, 2: review, 3: relearning)
	State int `json:"state" example:"0"`

	// Interval until next review (in days)
	Interval int `json:"interval" example:"1"`

	// Ease factor (in per mille, e.g., 2500 for 250%)
	Ease int `json:"ease" example:"2500"`

	// Number of reviews performed
	Reviews int `json:"reviews" example:"0"`

	// Number of lapses
	Lapses int `json:"lapses" example:"0"`

	// Next review date (Unix timestamp in seconds)
	Due int64 `json:"due" example:"1705324200"`

	// Ordinal position for display
	Ord int `json:"ord" example:"0"`

	// Colored flag (0: none, 1-7: colored flags)
	Flags int `json:"flags" example:"0"`

	// Whether the card is suspended
	Suspended bool `json:"suspended" example:"false"`
}

// SetCardFlagRequest represents the request payload to set a flag on a card
type SetCardFlagRequest struct {
	// Flag number (0: none, 1-7: colored flags)
	Flag int `json:"flag" example:"1" validate:"min=0,max=7"`
}

// ListCardsRequest represents the query parameters for listing cards
type ListCardsRequest struct {
	// Filter by deck ID
	DeckID *int64 `query:"deck_id"`

	// Filter by state (new, learn, review, relearn)
	State *string `query:"state" validate:"omitempty,oneof=new learn review relearn"`

	// Filter by flag (0-7)
	Flag *int `query:"flag" validate:"omitempty,min=0,max=7"`

	// Filter by suspended
	Suspended *bool `query:"suspended"`

	// Filter by buried
	Buried *bool `query:"buried"`

	// Page number (default: 1)
	Page int `query:"page" validate:"omitempty,min=1"`

	// Items per page (default: 20, max: 100)
	Limit int `query:"limit" validate:"omitempty,min=1,max=100"`
}

