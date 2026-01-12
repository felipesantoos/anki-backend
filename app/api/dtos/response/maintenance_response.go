package response

// EmptyCardsResponse represents the response payload for listing empty cards
// @Description Response payload containing a list of cards that render to empty
type EmptyCardsResponse struct {
	// Total number of empty cards found
	Count int `json:"count" example:"5"`
	// List of empty cards
	Data []*CardResponse `json:"data"`
}

// CleanupEmptyCardsResponse represents the response payload for cleaning up empty cards
// @Description Response payload containing the number of deleted empty cards
type CleanupEmptyCardsResponse struct {
	// Number of empty cards that were deleted
	DeletedCount int `json:"deleted_count" example:"5"`
}
