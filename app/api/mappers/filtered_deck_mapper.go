package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
)

// ToFilteredDeckResponse converts a FilteredDeck domain entity to a FilteredDeckResponse DTO
func ToFilteredDeckResponse(fd *filtereddeck.FilteredDeck) *response.FilteredDeckResponse {
	if fd == nil {
		return nil
	}
	return &response.FilteredDeckResponse{
		ID:           fd.GetID(),
		UserID:       fd.GetUserID(),
		Name:         fd.GetName(),
		SearchFilter: fd.GetSearchFilter(),
		Limit:        fd.GetLimitCards(),
		OrderBy:      fd.GetOrderBy(),
		Reschedule:   fd.GetReschedule(),
		CreatedAt:    fd.GetCreatedAt(),
		UpdatedAt:    fd.GetUpdatedAt(),
	}
}

// ToFilteredDeckResponseList converts a list of FilteredDeck domain entities to a list of FilteredDeckResponse DTOs
func ToFilteredDeckResponseList(decks []*filtereddeck.FilteredDeck) []*response.FilteredDeckResponse {
	res := make([]*response.FilteredDeckResponse, len(decks))
	for i, fd := range decks {
		res[i] = ToFilteredDeckResponse(fd)
	}
	return res
}

