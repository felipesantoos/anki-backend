package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestDeckToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	parentID := int64(10)
	deletedAt := now.Add(time.Hour)

	model := &models.DeckModel{
		ID:          1,
		UserID:      100,
		Name:        "Test Deck",
		ParentID:    sqlNullInt64(parentID, true),
		OptionsJSON: `{"newCardsPerDay":20}`,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   sqlNullTime(deletedAt, true),
	}

	entity, err := DeckToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "Test Deck", entity.GetName())
	assert.NotNil(t, entity.GetParentID())
	assert.Equal(t, parentID, *entity.GetParentID())
	assert.Equal(t, `{"newCardsPerDay":20}`, entity.GetOptionsJSON())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now, entity.GetUpdatedAt())
	assert.NotNil(t, entity.GetDeletedAt())
	assert.Equal(t, deletedAt, *entity.GetDeletedAt())
}

func TestDeckToDomain_WithNullFields(t *testing.T) {
	now := time.Now()

	model := &models.DeckModel{
		ID:          2,
		UserID:      200,
		Name:        "Root Deck",
		ParentID:    sqlNullInt64(0, false),
		OptionsJSON: "{}",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   sqlNullTime(time.Time{}, false),
	}

	entity, err := DeckToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetParentID())
	assert.Nil(t, entity.GetDeletedAt())
}

func TestDeckToDomain_NilInput(t *testing.T) {
	entity, err := DeckToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestDeckToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	parentID := int64(10)
	deletedAt := now.Add(time.Hour)

	entity, _ := DeckToDomain(&models.DeckModel{
		ID:          1,
		UserID:      100,
		Name:        "Test Deck",
		ParentID:    sqlNullInt64(parentID, true),
		OptionsJSON: `{"newCardsPerDay":20}`,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   sqlNullTime(deletedAt, true),
	})

	model := DeckToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "Test Deck", model.Name)
	assert.True(t, model.ParentID.Valid)
	assert.Equal(t, parentID, model.ParentID.Int64)
	assert.Equal(t, `{"newCardsPerDay":20}`, model.OptionsJSON)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestDeckToModel_WithNullFields(t *testing.T) {
	entity, _ := DeckToDomain(&models.DeckModel{
		ID:          2,
		UserID:      200,
		Name:        "Root Deck",
		ParentID:    sqlNullInt64(0, false),
		OptionsJSON: "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   sqlNullTime(time.Time{}, false),
	})

	model := DeckToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.ParentID.Valid)
	assert.False(t, model.DeletedAt.Valid)
}

