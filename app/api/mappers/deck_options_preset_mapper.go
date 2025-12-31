package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
)

// ToDeckOptionsPresetResponse converts a DeckOptionsPreset domain entity to a DeckOptionsPresetResponse DTO
func ToDeckOptionsPresetResponse(p *deckoptionspreset.DeckOptionsPreset) *response.DeckOptionsPresetResponse {
	if p == nil {
		return nil
	}
	return &response.DeckOptionsPresetResponse{
		ID:          p.GetID(),
		UserID:      p.GetUserID(),
		Name:        p.GetName(),
		OptionsJSON: p.GetOptionsJSON(),
		CreatedAt:   p.GetCreatedAt(),
		UpdatedAt:   p.GetUpdatedAt(),
	}
}

// ToDeckOptionsPresetResponseList converts a list of DeckOptionsPreset domain entities to a list of DeckOptionsPresetResponse DTOs
func ToDeckOptionsPresetResponseList(presets []*deckoptionspreset.DeckOptionsPreset) []*response.DeckOptionsPresetResponse {
	res := make([]*response.DeckOptionsPresetResponse, len(presets))
	for i, p := range presets {
		res[i] = ToDeckOptionsPresetResponse(p)
	}
	return res
}

