package response

// SearchResult represents the response payload for advanced search
type SearchResult struct {
	// Search results (NoteResponse or CardResponse)
	Data []interface{} `json:"data"`

	// Total number of results (for pagination)
	Total int `json:"total" example:"10"`
}

