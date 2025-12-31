package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
)

// ToDeckResponse converts a Deck domain entity to a DeckResponse DTO
func ToDeckResponse(d *deck.Deck) *response.DeckResponse {
	return ToDeckResponseWithContext(d, nil)
}

// ToDeckResponseWithContext converts a Deck domain entity to a DeckResponse DTO, 
// using the provided decks to calculate the full path name.
func ToDeckResponseWithContext(d *deck.Deck, allDecks []*deck.Deck) *response.DeckResponse {
	if d == nil {
		return nil
	}

	fullName := d.GetName()
	if allDecks != nil {
		fullName = d.GetFullPath(allDecks)
	}

	return &response.DeckResponse{
		ID:          d.GetID(),
		UserID:      d.GetUserID(),
		Name:        d.GetName(),
		FullName:    fullName,
		ParentID:    d.GetParentID(),
		OptionsJSON: d.GetOptionsJSON(),
		CreatedAt:   d.GetCreatedAt(),
		UpdatedAt:   d.GetUpdatedAt(),
	}
}

// ToDeckResponseList converts a list of Deck domain entities to a list of DeckResponse DTOs
func ToDeckResponseList(decks []*deck.Deck) []*response.DeckResponse {
	res := make([]*response.DeckResponse, len(decks))
	for i, d := range decks {
		res[i] = ToDeckResponseWithContext(d, decks)
	}
	return res
}

