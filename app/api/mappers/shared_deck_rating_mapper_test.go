package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestToSharedDeckRatingResponse(t *testing.T) {
	now := time.Now()
	comment := "nice"
	ratingVO, _ := valueobjects.NewSharedDeckRating(5)
	
	r, _ := shareddeckrating.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithSharedDeckID(2).
		WithRating(ratingVO).
		WithComment(&comment).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToSharedDeckRatingResponse(r)
		assert.NotNil(t, res)
		assert.Equal(t, r.GetID(), res.ID)
		assert.Equal(t, r.GetUserID(), res.UserID)
		assert.Equal(t, r.GetSharedDeckID(), res.SharedDeckID)
		assert.Equal(t, r.GetRating().Value(), res.Rating)
		assert.Equal(t, r.GetComment(), res.Comment)
		assert.Equal(t, r.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, r.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToSharedDeckRatingResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToSharedDeckRatingResponseList(t *testing.T) {
	ratingVO, _ := valueobjects.NewSharedDeckRating(5)
	r1, _ := shareddeckrating.NewBuilder().WithID(1).WithUserID(1).WithSharedDeckID(1).WithRating(ratingVO).Build()
	r2, _ := shareddeckrating.NewBuilder().WithID(2).WithUserID(1).WithSharedDeckID(1).WithRating(ratingVO).Build()
	ratings := []*shareddeckrating.SharedDeckRating{r1, r2}

	res := ToSharedDeckRatingResponseList(ratings)
	assert.Len(t, res, 2)
	assert.Equal(t, r1.GetID(), res[0].ID)
	assert.Equal(t, r2.GetID(), res[1].ID)
}
