package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
)

// ToDeckStatsResponse converts a DeckStats domain entity to a DeckStatsResponse DTO
func ToDeckStatsResponse(s *deck.DeckStats) *response.DeckStatsResponse {
	if s == nil {
		return nil
	}
	return &response.DeckStatsResponse{
		DeckID:         s.DeckID,
		NewCount:       s.NewCount,
		LearningCount:  s.LearningCount,
		ReviewCount:    s.ReviewCount,
		SuspendedCount: s.SuspendedCount,
		NotesCount:     s.NotesCount,
		DueTodayCount:  s.DueTodayCount,
	}
}

