package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestDeckModel_Creation(t *testing.T) {
	now := time.Now()
	parentID := int64(10)
	deletedAt := now.Add(time.Hour)

	model := &models.DeckModel{
		ID:          1,
		UserID:      100,
		Name:        "Test Deck",
		ParentID:    sqlNullInt64(parentID, true),
		OptionsJSON: "{}",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   sqlNullTime(deletedAt, true),
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "Test Deck", model.Name)
	assert.True(t, model.ParentID.Valid)
	assert.Equal(t, parentID, model.ParentID.Int64)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestDeckModel_NullFields(t *testing.T) {
	model := &models.DeckModel{
		ID:        2,
		UserID:    200,
		Name:      "Root Deck",
		ParentID:  sqlNullInt64(0, false),
		DeletedAt: sqlNullTime(time.Time{}, false),
	}

	assert.False(t, model.ParentID.Valid)
	assert.False(t, model.DeletedAt.Valid)
}

