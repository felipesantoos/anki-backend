package search

import (
	"context"
	"fmt"

	searchdomain "github.com/felipesantos/anki-backend/core/domain/services/search"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// SearchService implements ISearchService
type SearchService struct {
	noteRepo     secondary.INoteRepository
	cardRepo     secondary.ICardRepository
	parser       *searchdomain.Parser
}

// NewSearchService creates a new SearchService instance
func NewSearchService(
	noteRepo secondary.INoteRepository,
	cardRepo secondary.ICardRepository,
) primary.ISearchService {
	return &SearchService{
		noteRepo: noteRepo,
		cardRepo: cardRepo,
		parser:   searchdomain.NewParser(),
	}
}

// SearchAdvanced performs advanced search using Anki syntax
func (s *SearchService) SearchAdvanced(ctx context.Context, userID int64, query string, resultType string, limit int, offset int) (*primary.SearchResult, error) {
	// Set defaults
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Validate result type
	if resultType != "notes" && resultType != "cards" {
		return nil, fmt.Errorf("invalid result type: %s (must be 'notes' or 'cards')", resultType)
	}

	// Parse query
	parsedQuery, err := s.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Check if query has card-specific filters (states, flags, properties)
	hasCardFilters := len(parsedQuery.States) > 0 || len(parsedQuery.Flags) > 0 || len(parsedQuery.PropertyFilters) > 0

	var results []interface{}

	if resultType == "cards" || hasCardFilters {
		// Query cards first if we need card-specific filters or if result type is cards
		cards, err := s.cardRepo.FindByAdvancedSearch(ctx, userID, parsedQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to find cards: %w", err)
		}

		if resultType == "cards" {
			// Return cards directly
			for _, c := range cards {
				results = append(results, c)
			}
		} else {
			// Get note IDs from cards that match card filters
			noteIDs := make(map[int64]bool)
			for _, c := range cards {
				noteIDs[c.GetNoteID()] = true
			}

			// Query notes with note filters
			allNotes, err := s.noteRepo.FindByAdvancedSearch(ctx, userID, parsedQuery, 10000, 0) // Large limit to get all matching notes
			if err != nil {
				return nil, fmt.Errorf("failed to find notes: %w", err)
			}

			// If we have card filters, intersect with card results
			// Otherwise, use all notes from note query
			if hasCardFilters {
				// Filter notes that have matching cards
				for _, n := range allNotes {
					if noteIDs[n.GetID()] {
						results = append(results, n)
					}
				}
			} else {
				// No card filters, use all notes from query
				for _, n := range allNotes {
					results = append(results, n)
				}
			}
		}
	} else {
		// Query notes directly
		notes, err := s.noteRepo.FindByAdvancedSearch(ctx, userID, parsedQuery, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to find notes: %w", err)
		}
		for _, n := range notes {
			results = append(results, n)
		}
	}

	// Apply pagination if not already applied
	if hasCardFilters && resultType == "notes" {
		// Apply pagination manually
		start := offset
		end := offset + limit
		if start > len(results) {
			results = []interface{}{}
		} else {
			if end > len(results) {
				end = len(results)
			}
			results = results[start:end]
		}
	}

	return &primary.SearchResult{
		Data:  results,
		Total: len(results), // Note: This is approximate, full count would require separate query
	}, nil
}

