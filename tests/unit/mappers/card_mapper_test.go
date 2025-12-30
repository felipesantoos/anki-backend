package mappers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestCardToDomain_WithAllFields(t *testing.T) {
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
		State:        valueobjects.CardStateNew.String(),
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

	entity, err := CardToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetNoteID())
	assert.Equal(t, 2, entity.GetCardTypeID())
	assert.Equal(t, int64(5), entity.GetDeckID())
	assert.NotNil(t, entity.GetHomeDeckID())
	assert.Equal(t, homeDeckID, *entity.GetHomeDeckID())
	assert.Equal(t, int64(1234567890), entity.GetDue())
	assert.Equal(t, 86400, entity.GetInterval())
	assert.Equal(t, 2500, entity.GetEase())
	assert.Equal(t, 0, entity.GetLapses())
	assert.Equal(t, 5, entity.GetReps())
	assert.Equal(t, valueobjects.CardStateNew, entity.GetState())
	assert.Equal(t, 1, entity.GetPosition())
	assert.Equal(t, 0, entity.GetFlag())
	assert.False(t, entity.GetSuspended())
	assert.False(t, entity.GetBuried())
	assert.NotNil(t, entity.GetStability())
	assert.Equal(t, stability, *entity.GetStability())
	assert.NotNil(t, entity.GetDifficulty())
	assert.Equal(t, difficulty, *entity.GetDifficulty())
	assert.NotNil(t, entity.GetLastReviewAt())
	assert.Equal(t, lastReviewAt, *entity.GetLastReviewAt())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
}

func TestCardToDomain_WithNullFields(t *testing.T) {
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
		State:        valueobjects.CardStateLearning.String(),
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

	entity, err := CardToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetHomeDeckID())
	assert.Nil(t, entity.GetStability())
	assert.Nil(t, entity.GetDifficulty())
	assert.Nil(t, entity.GetLastReviewAt())
	assert.True(t, entity.GetSuspended())
	assert.True(t, entity.GetBuried())
}

func TestCardToDomain_NilInput(t *testing.T) {
	entity, err := CardToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestCardToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	homeDeckID := int64(10)
	stability := 0.85
	difficulty := 0.75
	lastReviewAt := now.Add(-time.Hour)

	entity, err := card.NewBuilder().
		WithID(1).
		WithNoteID(100).
		WithCardTypeID(2).
		WithDeckID(5).
		WithHomeDeckID(&homeDeckID).
		WithDue(1234567890).
		WithInterval(86400).
		WithEase(2500).
		WithLapses(0).
		WithReps(5).
		WithState(valueobjects.CardStateNew).
		WithPosition(1).
		WithFlag(0).
		WithSuspended(false).
		WithBuried(false).
		WithStability(&stability).
		WithDifficulty(&difficulty).
		WithLastReviewAt(&lastReviewAt).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		Build()
	require.NoError(t, err)

	model := CardToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.NoteID)
	assert.Equal(t, 2, model.CardTypeID)
	assert.Equal(t, int64(5), model.DeckID)
	assert.True(t, model.HomeDeckID.Valid)
	assert.Equal(t, homeDeckID, model.HomeDeckID.Int64)
	assert.Equal(t, int64(1234567890), model.Due)
	assert.Equal(t, 86400, model.Interval)
	assert.Equal(t, 2500, model.Ease)
	assert.Equal(t, 0, model.Lapses)
	assert.Equal(t, 5, model.Reps)
	assert.Equal(t, valueobjects.CardStateNew.String(), model.State)
	assert.Equal(t, 1, model.Position)
	assert.Equal(t, 0, model.Flag)
	assert.False(t, model.Suspended)
	assert.False(t, model.Buried)
	assert.True(t, model.Stability.Valid)
	assert.Equal(t, stability, model.Stability.Float64)
	assert.True(t, model.Difficulty.Valid)
	assert.Equal(t, difficulty, model.Difficulty.Float64)
	assert.True(t, model.LastReviewAt.Valid)
	assert.Equal(t, lastReviewAt, model.LastReviewAt.Time)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
}

func TestCardToModel_WithNullFields(t *testing.T) {
	now := time.Now()

	entity, err := card.NewBuilder().
		WithID(2).
		WithNoteID(200).
		WithCardTypeID(1).
		WithDeckID(10).
		WithHomeDeckID(nil).
		WithDue(9876543210).
		WithInterval(3600).
		WithEase(2000).
		WithLapses(1).
		WithReps(3).
		WithState(valueobjects.CardStateLearning).
		WithPosition(0).
		WithFlag(1).
		WithSuspended(true).
		WithBuried(true).
		WithStability(nil).
		WithDifficulty(nil).
		WithLastReviewAt(nil).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	require.NoError(t, err)

	model := CardToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.HomeDeckID.Valid)
	assert.False(t, model.Stability.Valid)
	assert.False(t, model.Difficulty.Valid)
	assert.False(t, model.LastReviewAt.Valid)
}

func TestCardToDomain_AllCardStates(t *testing.T) {
	states := []valueobjects.CardState{
		valueobjects.CardStateNew,
		valueobjects.CardStateLearning,
		valueobjects.CardStateReview,
		valueobjects.CardStateRelearning,
	}

	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			model := &models.CardModel{
				ID:         1,
				NoteID:     100,
				CardTypeID: 1,
				DeckID:     5,
				Due:        1234567890,
				Interval:   86400,
				Ease:       2500,
				Lapses:     0,
				Reps:       0,
				State:      state.String(),
				Position:   0,
				Flag:       0,
				Suspended:  false,
				Buried:     false,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			entity, err := CardToDomain(model)
			require.NoError(t, err)
			assert.Equal(t, state, entity.GetState())
		})
	}
}

