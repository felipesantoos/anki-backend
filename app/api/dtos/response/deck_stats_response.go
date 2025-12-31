package response

// DeckStatsResponse represents the study statistics for a deck
// @Description Response payload containing deck study statistics
type DeckStatsResponse struct {
	// Unique identifier for the deck
	DeckID int64 `json:"deck_id" example:"1"`

	// Number of new cards
	NewCount int `json:"new_count" example:"50"`

	// Number of cards in learning/relearning state
	LearningCount int `json:"learning_count" example:"10"`

	// Number of cards in review state
	ReviewCount int `json:"review_count" example:"100"`

	// Number of suspended cards
	SuspendedCount int `json:"suspended_count" example:"5"`

	// Total number of notes associated with the deck
	NotesCount int `json:"notes_count" example:"150"`

	// Number of cards due today
	DueTodayCount int `json:"due_today_count" example:"25"`
}

