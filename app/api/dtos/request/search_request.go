package request

// AdvancedSearchRequest represents the request payload for advanced search
type AdvancedSearchRequest struct {
	// Anki search query string (e.g., "deck:Default tag:vocabulary front:hello -tag:marked")
	Query string `json:"query" example:"deck:Default tag:vocabulary front:hello -tag:marked" validate:"required"`

	// Result type: "notes" or "cards"
	Type string `json:"type" example:"notes" validate:"required,oneof=notes cards"`

	// Maximum number of results
	Limit int `json:"limit" example:"20"`

	// Pagination offset
	Offset int `json:"offset" example:"0"`
}

