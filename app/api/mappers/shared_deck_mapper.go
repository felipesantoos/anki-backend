package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
)

// ToSharedDeckResponse converts SharedDeck domain entity to Response DTO
func ToSharedDeckResponse(sd *shareddeck.SharedDeck) *response.SharedDeckResponse {
	if sd == nil {
		return nil
	}
	return &response.SharedDeckResponse{
		ID:            sd.GetID(),
		AuthorID:      sd.GetAuthorID(),
		Name:          sd.GetName(),
		Description:   sd.GetDescription(),
		Category:      sd.GetCategory(),
		PackagePath:   sd.GetPackagePath(),
		PackageSize:   sd.GetPackageSize(),
		DownloadCount: sd.GetDownloadCount(),
		IsPublic:      sd.GetIsPublic(),
		Tags:          sd.GetTags(),
		CreatedAt:     sd.GetCreatedAt(),
		UpdatedAt:     sd.GetUpdatedAt(),
	}
}

// ToSharedDeckResponseList converts list of SharedDeck domain entities to list of Response DTOs
func ToSharedDeckResponseList(decks []*shareddeck.SharedDeck) []*response.SharedDeckResponse {
	res := make([]*response.SharedDeckResponse, len(decks))
	for i, sd := range decks {
		res[i] = ToSharedDeckResponse(sd)
	}
	return res
}

