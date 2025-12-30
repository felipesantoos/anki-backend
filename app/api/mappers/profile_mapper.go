package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
)

// ToProfileResponse converts a Profile domain entity to a ProfileResponse DTO
func ToProfileResponse(p *profile.Profile) *response.ProfileResponse {
	if p == nil {
		return nil
	}
	return &response.ProfileResponse{
		ID:           p.GetID(),
		UserID:       p.GetUserID(),
		Name:         p.GetName(),
		SyncEnabled:  p.GetAnkiWebSyncEnabled(),
		SyncUsername: p.GetAnkiWebUsername(),
		CreatedAt:    p.GetCreatedAt(),
		UpdatedAt:    p.GetUpdatedAt(),
	}
}

// ToProfileResponseList converts a list of Profile domain entities to a list of ProfileResponse DTOs
func ToProfileResponseList(profiles []*profile.Profile) []*response.ProfileResponse {
	res := make([]*response.ProfileResponse, len(profiles))
	for i, p := range profiles {
		res[i] = ToProfileResponse(p)
	}
	return res
}

