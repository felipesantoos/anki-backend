package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestCardModel_Creation(t *testing.T) {
	now := time.Now()
	homeDeckID := int64(10)
	stability := 0.85
	difficulty := 0.75
	lastReviewAt := now.Add(-time.Hour)

	model := &models.CardModel{
		ID:           1,
		NoteID:       100,
		CardTypeID:   2,
		DeckID:       5,
		HomeDeckID:   sqlNullInt64(homeDeckID, true),
		Due:          1234567890,
		Interval:     86400,
		Ease:         2500,
		Lapses:       0,
		Reps:         5,
		State:        "new",
		Position:     1,
		Flag:         0,
		Suspended:    false,
		Buried:       false,
		Stability:    sqlNullFloat64(stability, true),
		Difficulty:   sqlNullFloat64(difficulty, true),
		LastReviewAt: sqlNullTime(lastReviewAt, true),
		CreatedAt:    now,
		UpdatedAt:    now.Add(time.Hour),
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.NoteID)
	assert.Equal(t, 2, model.CardTypeID)
	assert.True(t, model.HomeDeckID.Valid)
	assert.Equal(t, homeDeckID, model.HomeDeckID.Int64)
	assert.True(t, model.Stability.Valid)
	assert.Equal(t, stability, model.Stability.Float64)
	assert.True(t, model.Difficulty.Valid)
	assert.Equal(t, difficulty, model.Difficulty.Float64)
	assert.True(t, model.LastReviewAt.Valid)
	assert.Equal(t, lastReviewAt, model.LastReviewAt.Time)
}

func TestCardModel_NullFields(t *testing.T) {
	now := time.Now()

	model := &models.CardModel{
		ID:           2,
		NoteID:       200,
		CardTypeID:   1,
		DeckID:       10,
		HomeDeckID:   sqlNullInt64(0, false),
		Due:          9876543210,
		Interval:     3600,
		Ease:         2000,
		Lapses:       1,
		Reps:         3,
		State:        "learn",
		Position:     0,
		Flag:         1,
		Suspended:    true,
		Buried:       true,
		Stability:    sqlNullFloat64(0, false),
		Difficulty:   sqlNullFloat64(0, false),
		LastReviewAt: sqlNullTime(time.Time{}, false),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.False(t, model.HomeDeckID.Valid)
	assert.False(t, model.Stability.Valid)
	assert.False(t, model.Difficulty.Valid)
	assert.False(t, model.LastReviewAt.Valid)
}

