package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	"github.com/stretchr/testify/assert"
)

func TestToReviewResponse(t *testing.T) {
	now := time.Now()
	
	r, _ := review.NewBuilder().
		WithID(10).
		WithCardID(1).
		WithRating(3).
		WithTimeMs(5000).
		WithCreatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToReviewResponse(r)
		assert.NotNil(t, res)
		assert.Equal(t, r.GetID(), res.ID)
		assert.Equal(t, r.GetCardID(), res.CardID)
		assert.Equal(t, int(r.GetRating()), res.Rating)
		assert.Equal(t, r.GetTimeMs(), res.TimeMs)
		assert.Equal(t, r.GetCreatedAt(), res.CreatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToReviewResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToReviewResponseList(t *testing.T) {
	r1, _ := review.NewBuilder().WithID(1).Build()
	r2, _ := review.NewBuilder().WithID(2).Build()
	reviews := []*review.Review{r1, r2}

	res := ToReviewResponseList(reviews)
	assert.Len(t, res, 2)
	assert.Equal(t, r1.GetID(), res[0].ID)
	assert.Equal(t, r2.GetID(), res[1].ID)
}
