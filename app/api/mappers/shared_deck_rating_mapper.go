package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
)

// ToSharedDeckRatingResponse converts SharedDeckRating domain entity to Response DTO
func ToSharedDeckRatingResponse(r *shareddeckrating.SharedDeckRating) *response.SharedDeckRatingResponse {
	if r == nil {
		return nil
	}
	return &response.SharedDeckRatingResponse{
		ID:           r.GetID(),
		UserID:       r.GetUserID(),
		SharedDeckID: r.GetSharedDeckID(),
		Rating:       r.GetRating().Value(),
		Comment:      r.GetComment(),
		CreatedAt:    r.GetCreatedAt(),
		UpdatedAt:    r.GetUpdatedAt(),
	}
}

// ToSharedDeckRatingResponseList converts list of SharedDeckRating domain entities to list of Response DTOs
func ToSharedDeckRatingResponseList(ratings []*shareddeckrating.SharedDeckRating) []*response.SharedDeckRatingResponse {
	res := make([]*response.SharedDeckRatingResponse, len(ratings))
	for i, r := range ratings {
		res[i] = ToSharedDeckRatingResponse(r)
	}
	return res
}

