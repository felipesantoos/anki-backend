package primary

import (
	"context"
)

// SearchResult represents the result of an advanced search
type SearchResult struct {
	Data  []interface{} `json:"data"`  // NoteResponse or CardResponse
	Total int           `json:"total"` // Total count (for pagination)
}

// ISearchService defines the interface for advanced search functionality
type ISearchService interface {
	// SearchAdvanced performs advanced search using Anki syntax
	// query: Anki search query string (e.g., "deck:Default tag:vocabulary front:hello")
	// resultType: "notes" or "cards"
	// limit: Maximum number of results
	// offset: Pagination offset
	SearchAdvanced(ctx context.Context, userID int64, query string, resultType string, limit int, offset int) (*SearchResult, error)
}

