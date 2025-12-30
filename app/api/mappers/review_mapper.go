package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/review"
)

// ToReviewResponse converts a Review domain entity to a ReviewResponse DTO
func ToReviewResponse(r *review.Review) *response.ReviewResponse {
	if r == nil {
		return nil
	}
	return &response.ReviewResponse{
		ID:        r.GetID(),
		CardID:    r.GetCardID(),
		Rating:    int(r.GetRating()),
		TimeMs:    r.GetTimeMs(),
		CreatedAt: r.GetCreatedAt(),
	}
}

// ToReviewResponseList converts a list of Review domain entities to a list of ReviewResponse DTOs
func ToReviewResponseList(reviews []*review.Review) []*response.ReviewResponse {
	res := make([]*response.ReviewResponse, len(reviews))
	for i, r := range reviews {
		res[i] = ToReviewResponse(r)
	}
	return res
}

