package mappers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestNoteTypeToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	model := &models.NoteTypeModel{
		ID:            1,
		UserID:        100,
		Name:          "Basic",
		FieldsJSON:    `[{"name":"Front"},{"name":"Back"}]`,
		CardTypesJSON: `[{"name":"Card 1"}]`,
		TemplatesJSON: `[{"name":"Card 1 Front"}]`,
		CreatedAt:     now,
		UpdatedAt:     now.Add(time.Hour),
		DeletedAt:     sqlNullTime(deletedAt, true),
	}

	entity, err := NoteTypeToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "Basic", entity.GetName())
	assert.Equal(t, `[{"name":"Front"},{"name":"Back"}]`, entity.GetFieldsJSON())
	assert.Equal(t, `[{"name":"Card 1"}]`, entity.GetCardTypesJSON())
	assert.Equal(t, `[{"name":"Card 1 Front"}]`, entity.GetTemplatesJSON())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
	assert.NotNil(t, entity.GetDeletedAt())
	assert.Equal(t, deletedAt, *entity.GetDeletedAt())
}

func TestNoteTypeToDomain_WithNullDeletedAt(t *testing.T) {
	now := time.Now()

	model := &models.NoteTypeModel{
		ID:            2,
		UserID:        200,
		Name:          "Cloze",
		FieldsJSON:    `[{"name":"Text"}]`,
		CardTypesJSON: `[{"name":"Cloze"}]`,
		TemplatesJSON: `[{"name":"Cloze Template"}]`,
		CreatedAt:     now,
		UpdatedAt:     now,
		DeletedAt:     sqlNullTime(time.Time{}, false),
	}

	entity, err := NoteTypeToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetDeletedAt())
}

func TestNoteTypeToDomain_NilInput(t *testing.T) {
	entity, err := NoteTypeToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestNoteTypeToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	entity, err := notetype.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Card 1 Front"}]`).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		WithDeletedAt(&deletedAt).
		Build()
	require.NoError(t, err)

	model := NoteTypeToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "Basic", model.Name)
	assert.Equal(t, `[{"name":"Front"},{"name":"Back"}]`, model.FieldsJSON)
	assert.Equal(t, `[{"name":"Card 1"}]`, model.CardTypesJSON)
	assert.Equal(t, `[{"name":"Card 1 Front"}]`, model.TemplatesJSON)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestNoteTypeToModel_WithNullDeletedAt(t *testing.T) {
	now := time.Now()

	entity, err := notetype.NewBuilder().
		WithID(2).
		WithUserID(200).
		WithName("Cloze").
		WithFieldsJSON(`[{"name":"Text"}]`).
		WithCardTypesJSON(`[{"name":"Cloze"}]`).
		WithTemplatesJSON(`[{"name":"Cloze Template"}]`).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		WithDeletedAt(nil).
		Build()
	require.NoError(t, err)

	model := NoteTypeToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.DeletedAt.Valid)
}

