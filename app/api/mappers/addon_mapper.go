package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/add_on"
)

// ToAddOnResponse converts AddOn domain entity to Response DTO
func ToAddOnResponse(a *addon.AddOn) *response.AddOnResponse {
	if a == nil {
		return nil
	}
	return &response.AddOnResponse{
		ID:         a.GetID(),
		UserID:     a.GetUserID(),
		Code:       a.GetCode(),
		Name:       a.GetName(),
		Version:    a.GetVersion(),
		ConfigJSON: a.GetConfigJSON(),
		Enabled:    a.GetEnabled(),
		CreatedAt:  a.GetInstalledAt(),
		UpdatedAt:  a.GetUpdatedAt(),
	}
}

// ToAddOnResponseList converts list of AddOn domain entities to list of Response DTOs
func ToAddOnResponseList(addOns []*addon.AddOn) []*response.AddOnResponse {
	res := make([]*response.AddOnResponse, len(addOns))
	for i, a := range addOns {
		res[i] = ToAddOnResponse(a)
	}
	return res
}

