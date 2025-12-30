package mappers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestNoteToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	guid, err := valueobjects.NewGUID("test-guid-123")
	require.NoError(t, err)

	model := &models.NoteModel{
		ID:         1,
		UserID:     100,
		GUID:       guid.Value(),
		NoteTypeID: 5,
		FieldsJSON: `{"field1":"value1","field2":"value2"}`,
		Tags:       sqlNullString("tag1,tag2,tag3", true),
		Marked:     true,
		CreatedAt:  now,
		UpdatedAt:  now.Add(time.Hour),
		DeletedAt:  sqlNullTime(now.Add(2*time.Hour), true),
	}

	entity, err := NoteToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, guid.Value(), entity.GetGUID().Value())
	assert.Equal(t, int64(5), entity.GetNoteTypeID())
	assert.Equal(t, `{"field1":"value1","field2":"value2"}`, entity.GetFieldsJSON())
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, entity.GetTags())
	assert.True(t, entity.GetMarked())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
	assert.NotNil(t, entity.GetDeletedAt())
	assert.Equal(t, now.Add(2*time.Hour), *entity.GetDeletedAt())
}

func TestNoteToDomain_WithNullFields(t *testing.T) {
	now := time.Now()
	guid, err := valueobjects.NewGUID("test-guid-456")
	require.NoError(t, err)

	model := &models.NoteModel{
		ID:         2,
		UserID:     200,
		GUID:       guid.Value(),
		NoteTypeID: 10,
		FieldsJSON: `{}`,
		Tags:       sqlNullString("", false), // Null tags
		Marked:     false,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sqlNullTime(time.Time{}, false), // Null deleted_at
	}

	entity, err := NoteToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(2), entity.GetID())
	assert.Empty(t, entity.GetTags())
	assert.Nil(t, entity.GetDeletedAt())
}

func TestNoteToDomain_WithEmptyTagsString(t *testing.T) {
	now := time.Now()
	guid, err := valueobjects.NewGUID("test-guid-789")
	require.NoError(t, err)

	model := &models.NoteModel{
		ID:         3,
		UserID:     300,
		GUID:       guid.Value(),
		NoteTypeID: 15,
		FieldsJSON: `{"field":"value"}`,
		Tags:       sqlNullString("", true), // Empty string but valid
		Marked:     false,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sqlNullTime(time.Time{}, false),
	}

	entity, err := NoteToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Empty(t, entity.GetTags())
}

func TestNoteToDomain_NilInput(t *testing.T) {
	entity, err := NoteToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestNoteToDomain_InvalidGUID(t *testing.T) {
	model := &models.NoteModel{
		ID:         1,
		UserID:     100,
		GUID:       "", // Invalid GUID
		NoteTypeID: 5,
		FieldsJSON: `{}`,
		Tags:       sqlNullString("", false),
		Marked:     false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		DeletedAt:  sqlNullTime(time.Time{}, false),
	}

	entity, err := NoteToDomain(model)
	assert.Error(t, err)
	assert.Nil(t, entity)
}

func TestNoteToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	guid, err := valueobjects.NewGUID("test-guid-model")
	require.NoError(t, err)

	entity, err := note.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithGUID(guid).
		WithNoteTypeID(5).
		WithFieldsJSON(`{"field1":"value1"}`).
		WithTags([]string{"tag1", "tag2"}).
		WithMarked(true).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		WithDeletedAt(timePtr(now.Add(2 * time.Hour))).
		Build()
	require.NoError(t, err)

	model := NoteToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, guid.Value(), model.GUID)
	assert.Equal(t, int64(5), model.NoteTypeID)
	assert.Equal(t, `{"field1":"value1"}`, model.FieldsJSON)
	assert.True(t, model.Tags.Valid)
	assert.Contains(t, model.Tags.String, "tag1")
	assert.True(t, model.Marked)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, now.Add(2*time.Hour), model.DeletedAt.Time)
}

func TestNoteToModel_WithNullFields(t *testing.T) {
	now := time.Now()
	guid, err := valueobjects.NewGUID("test-guid-null")
	require.NoError(t, err)

	entity, err := note.NewBuilder().
		WithID(2).
		WithUserID(200).
		WithGUID(guid).
		WithNoteTypeID(10).
		WithFieldsJSON(`{}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		WithDeletedAt(nil).
		Build()
	require.NoError(t, err)

	model := NoteToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.Tags.Valid || model.Tags.String == "")
	assert.False(t, model.DeletedAt.Valid)
}

func TestNoteToModel_WithEmptyTags(t *testing.T) {
	now := time.Now()
	guid, err := valueobjects.NewGUID("test-guid-empty-tags")
	require.NoError(t, err)

	entity, err := note.NewBuilder().
		WithID(3).
		WithUserID(300).
		WithGUID(guid).
		WithNoteTypeID(15).
		WithFieldsJSON(`{}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	require.NoError(t, err)

	model := NoteToModel(entity)
	require.NotNil(t, model)

	// Empty tags should result in invalid or empty sql.NullString
	assert.False(t, model.Tags.Valid || len(model.Tags.String) == 0)
}

