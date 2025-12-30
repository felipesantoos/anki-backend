package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

)

func TestReviewModel_Creation(t *testing.T) {
	now := time.Now()

	model := &ReviewModel{
		ID:        1,
		CardID:    100,
		Rating:    3,
		Interval:  86400,
		Ease:      2500,
		TimeMs:    5000,
		Type:      "review",
		CreatedAt: now,
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.CardID)
	assert.Equal(t, 3, model.Rating)
	assert.Equal(t, 86400, model.Interval)
	assert.Equal(t, 2500, model.Ease)
	assert.Equal(t, 5000, model.TimeMs)
	assert.Equal(t, "review", model.Type)
	assert.Equal(t, now, model.CreatedAt)
}

func TestReviewModel_ZeroValues(t *testing.T) {
	model := &ReviewModel{}

	assert.Equal(t, int64(0), model.ID)
	assert.Equal(t, int64(0), model.CardID)
	assert.Equal(t, 0, model.Rating)
	assert.Equal(t, 0, model.Interval)
	assert.Equal(t, 0, model.Ease)
	assert.Equal(t, 0, model.TimeMs)
	assert.Equal(t, "", model.Type)
	assert.True(t, model.CreatedAt.IsZero())
}

