package mappers

import (
	"database/sql"

	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// SharedDeckRatingToDomain converts a SharedDeckRatingModel (database representation) to a SharedDeckRating entity (domain representation)
func SharedDeckRatingToDomain(model *models.SharedDeckRatingModel) (*shareddeckrating.SharedDeckRating, error) {
	if model == nil {
		return nil, nil
	}

	// Convert int to valueobjects.Rating (1-5 for shared decks)
	rating, err := valueobjects.NewSharedDeckRating(model.Rating)
	if err != nil {
		return nil, err
	}

	builder := shareddeckrating.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithSharedDeckID(model.SharedDeckID).
		WithRating(rating).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable comment
	if model.Comment.Valid {
		builder = builder.WithComment(&model.Comment.String)
	}

	return builder.Build()
}

// SharedDeckRatingToModel converts a SharedDeckRating entity (domain representation) to a SharedDeckRatingModel (database representation)
func SharedDeckRatingToModel(ratingEntity *shareddeckrating.SharedDeckRating) *models.SharedDeckRatingModel {
	model := &models.SharedDeckRatingModel{
		ID:           ratingEntity.GetID(),
		UserID:       ratingEntity.GetUserID(),
		SharedDeckID: ratingEntity.GetSharedDeckID(),
		Rating:       ratingEntity.GetRating().Value(), // Convert valueobjects.Rating to int
		CreatedAt:    ratingEntity.GetCreatedAt(),
		UpdatedAt:    ratingEntity.GetUpdatedAt(),
	}

	// Handle nullable comment
	if ratingEntity.GetComment() != nil {
		model.Comment = sql.NullString{
			String: *ratingEntity.GetComment(),
			Valid:  true,
		}
	}

	return model
}

