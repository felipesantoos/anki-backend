package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestNoteTypeModel_Creation(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	model := &models.NoteTypeModel{
		ID:            1,
		UserID:        100,
		Name:          "Basic",
		FieldsJSON:    `[{"name":"Front"}]`,
		CardTypesJSON: `[{"name":"Card 1"}]`,
		TemplatesJSON: `[{"name":"Template 1"}]`,
		CreatedAt:     now,
		UpdatedAt:     now.Add(time.Hour),
		DeletedAt:     sqlNullTime(deletedAt, true),
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Basic", model.Name)
	assert.True(t, model.DeletedAt.Valid)
}

func TestNoteTypeModel_NullDeletedAt(t *testing.T) {
	model := &models.NoteTypeModel{
		ID:    2,
		UserID: 200,
		Name:  "Cloze",
		DeletedAt: sqlNullTime(time.Time{}, false),
	}

	assert.False(t, model.DeletedAt.Valid)
}
