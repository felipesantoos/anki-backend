package mappers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestSharedDeckRatingToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	comment := "Great deck!"

	model := &models.SharedDeckRatingModel{
		ID:          1,
		SharedDeckID: 100,
		UserID:      200,
		Rating:      5,
		Comment:     sqlNullString(comment, true),
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Hour),
	}

	entity, err := SharedDeckRatingToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetSharedDeckID())
	assert.Equal(t, int64(200), entity.GetUserID())
	assert.Equal(t, valueobjects.Rating(5), entity.GetRating().Value())
	assert.NotNil(t, entity.GetComment())
	assert.Equal(t, comment, *entity.GetComment())
}

func TestSharedDeckRatingToDomain_WithNullComment(t *testing.T) {
	now := time.Now()

	model := &models.SharedDeckRatingModel{
		ID:          2,
		SharedDeckID: 100,
		UserID:      200,
		Rating:      4,
		Comment:     sqlNullString("", false),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	entity, err := SharedDeckRatingToDomain(model)
	require.NoError(t, err)
	assert.Nil(t, entity.GetComment())
}

func TestSharedDeckRatingToDomain_InvalidRating(t *testing.T) {
	now := time.Now()

	model := &models.SharedDeckRatingModel{
		ID:          3,
		SharedDeckID: 100,
		UserID:      200,
		Rating:      10, // Invalid (should be 1-5)
		Comment:     sqlNullString("", false),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	entity, err := SharedDeckRatingToDomain(model)
	assert.Error(t, err)
	assert.Nil(t, entity)
}

func TestSharedDeckRatingToDomain_NilInput(t *testing.T) {
	entity, err := SharedDeckRatingToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestSharedDeckRatingToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	comment := "Great deck!"
	rating, err := valueobjects.NewSharedDeckRating(5)
	require.NoError(t, err)

	entity, err := shareddeckrating.NewBuilder().
		WithID(1).
		WithUserID(200).
		WithSharedDeckID(100).
		WithRating(rating).
		WithComment(&comment).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		Build()
	require.NoError(t, err)

	model := SharedDeckRatingToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, 5, model.Rating)
	assert.True(t, model.Comment.Valid)
	assert.Equal(t, comment, model.Comment.String)
}

