package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
)

// ToCardResponse converts a Card domain entity to a CardResponse DTO
func ToCardResponse(c *card.Card) *response.CardResponse {
	if c == nil {
		return nil
	}
	return &response.CardResponse{
		ID:        c.GetID(),
		NoteID:    c.GetNoteID(),
		DeckID:    c.GetDeckID(),
		State:     string(c.GetState()),
		Interval:  c.GetInterval(),
		Ease:      c.GetEase(),
		Reviews:   c.GetReps(),
		Lapses:    c.GetLapses(),
		Due:       c.GetDue(),
		Ord:       c.GetPosition(),
		Flags:     c.GetFlag(),
		Suspended: c.GetSuspended(),
		CreatedAt: c.GetCreatedAt(),
		UpdatedAt: c.GetUpdatedAt(),
	}
}

// ToCardResponseList converts a list of Card domain entities to a list of CardResponse DTOs
func ToCardResponseList(cards []*card.Card) []*response.CardResponse {
	res := make([]*response.CardResponse, len(cards))
	for i, c := range cards {
		res[i] = ToCardResponse(c)
	}
	return res
}

