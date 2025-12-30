package mappers_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deckoptionspreset "github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestDeckOptionsPresetToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	model := &models.DeckOptionsPresetModel{
		ID:          1,
		UserID:      100,
		Name:        "Default Preset",
		OptionsJSON: `{"newCardsPerDay":20}`,
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Hour),
		DeletedAt:   sqlNullTime(deletedAt, true),
	}

	entity, err := mappers.DeckOptionsPresetToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "Default Preset", entity.GetName())
	assert.Equal(t, `{"newCardsPerDay":20}`, entity.GetOptionsJSON())
	assert.NotNil(t, entity.GetDeletedAt())
}

func TestDeckOptionsPresetToDomain_NilInput(t *testing.T) {
	entity, err := DeckOptionsPresetToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestDeckOptionsPresetToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	entity, err := deckoptionspreset.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithName("Default Preset").
		WithOptionsJSON(`{"newCardsPerDay":20}`).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		WithDeletedAt(&deletedAt).
		Build()
	require.NoError(t, err)

	model := mappers.DeckOptionsPresetToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Default Preset", model.Name)
	assert.True(t, model.DeletedAt.Valid)
}

