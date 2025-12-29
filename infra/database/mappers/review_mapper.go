package mappers

import (
	"fmt"

	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// ReviewToDomain converts a ReviewModel (database representation) to a Review entity (domain representation)
func ReviewToDomain(model *models.ReviewModel) (*review.Review, error) {
	if model == nil {
		return nil, nil
	}

	// Parse review type from string
	reviewType := valueobjects.ReviewType(model.Type)
	if !reviewType.IsValid() {
		return nil, fmt.Errorf("invalid review type: %s", model.Type)
	}

	builder := review.NewBuilder().
		WithID(model.ID).
		WithCardID(model.CardID).
		WithRating(model.Rating).
		WithInterval(model.Interval).
		WithEase(model.Ease).
		WithTimeMs(model.TimeMs).
		WithType(reviewType).
		WithCreatedAt(model.CreatedAt)

	return builder.Build()
}

// ReviewToModel converts a Review entity (domain representation) to a ReviewModel (database representation)
func ReviewToModel(reviewEntity *review.Review) *models.ReviewModel {
	return &models.ReviewModel{
		ID:        reviewEntity.GetID(),
		CardID:    reviewEntity.GetCardID(),
		Rating:    reviewEntity.GetRating(),
		Interval:  reviewEntity.GetInterval(),
		Ease:      reviewEntity.GetEase(),
		TimeMs:    reviewEntity.GetTimeMs(),
		Type:      reviewEntity.GetType().String(),
		CreatedAt: reviewEntity.GetCreatedAt(),
	}
}

