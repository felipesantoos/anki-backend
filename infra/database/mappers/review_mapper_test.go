package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestReviewToDomain_WithAllFields(t *testing.T) {
	now := time.Now()

	model := &models.ReviewModel{
		ID:         1,
		CardID:     100,
		Rating:     3,
		Interval:   86400,
		Ease:       2500,
		TimeMs:     5000,
		Type:       valueobjects.ReviewTypeReview.String(),
		CreatedAt:  now,
	}

	entity, err := ReviewToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetCardID())
	assert.Equal(t, 3, entity.GetRating())
	assert.Equal(t, 86400, entity.GetInterval())
	assert.Equal(t, 2500, entity.GetEase())
	assert.Equal(t, 5000, entity.GetTimeMs())
	assert.Equal(t, valueobjects.ReviewTypeReview, entity.GetType())
	assert.Equal(t, now, entity.GetCreatedAt())
}

func TestReviewToDomain_NilInput(t *testing.T) {
	entity, err := ReviewToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestReviewToModel_WithAllFields(t *testing.T) {
	now := time.Now()

	entity, err := review.NewBuilder().
		WithID(1).
		WithCardID(100).
		WithRating(3).
		WithInterval(86400).
		WithEase(2500).
		WithTimeMs(5000).
		WithType(valueobjects.ReviewTypeReview).
		WithCreatedAt(now).
		Build()
	require.NoError(t, err)

	model := ReviewToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.CardID)
	assert.Equal(t, 3, model.Rating)
	assert.Equal(t, 86400, model.Interval)
	assert.Equal(t, 2500, model.Ease)
	assert.Equal(t, 5000, model.TimeMs)
	assert.Equal(t, valueobjects.ReviewTypeReview.String(), model.Type)
	assert.Equal(t, now, model.CreatedAt)
}

func TestReviewToDomain_AllReviewTypes(t *testing.T) {
	reviewTypes := []valueobjects.ReviewType{
		valueobjects.ReviewTypeLearn,
		valueobjects.ReviewTypeReview,
		valueobjects.ReviewTypeRelearn,
		valueobjects.ReviewTypeCram,
	}

	for _, reviewType := range reviewTypes {
		t.Run(reviewType.String(), func(t *testing.T) {
			model := &models.ReviewModel{
				ID:         1,
				CardID:     100,
				Rating:     3,
				Interval:   86400,
				Ease:       2500,
				TimeMs:     5000,
				Type: reviewType.String(),
				CreatedAt:  time.Now(),
			}

			entity, err := ReviewToDomain(model)
			require.NoError(t, err)
			assert.Equal(t, reviewType, entity.GetType())
		})
	}
}

