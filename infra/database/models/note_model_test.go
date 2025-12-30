package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

)

func TestNoteModel_Creation(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	model := &NoteModel{
		ID:         1,
		UserID:     100,
		GUID:       "550e8400-e29b-41d4-a716-446655440000",
		NoteTypeID: 5,
		FieldsJSON: `{"field1":"value1"}`,
		Tags:       sqlNullString("{tag1,tag2}", true),
		Marked:     true,
		CreatedAt:  now,
		UpdatedAt:  now.Add(time.Hour),
		DeletedAt:  sqlNullTime(deletedAt, true),
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", model.GUID)
	assert.Equal(t, int64(5), model.NoteTypeID)
	assert.Equal(t, `{"field1":"value1"}`, model.FieldsJSON)
	assert.True(t, model.Tags.Valid)
	assert.Equal(t, "{tag1,tag2}", model.Tags.String)
	assert.True(t, model.Marked)
	assert.Equal(t, now, model.CreatedAt)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestNoteModel_NullFields(t *testing.T) {
	now := time.Now()

	model := &NoteModel{
		ID:         2,
		UserID:     200,
		GUID:       "550e8400-e29b-41d4-a716-446655440001",
		NoteTypeID: 10,
		FieldsJSON: `{}`,
		Tags:       sqlNullString("", false),
		Marked:     false,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sqlNullTime(time.Time{}, false),
	}

	assert.False(t, model.Tags.Valid)
	assert.False(t, model.DeletedAt.Valid)
}

func TestNoteModel_ZeroValues(t *testing.T) {
	model := &NoteModel{}

	assert.Equal(t, int64(0), model.ID)
	assert.Equal(t, int64(0), model.UserID)
	assert.Equal(t, "", model.GUID)
	assert.Equal(t, int64(0), model.NoteTypeID)
	assert.Equal(t, "", model.FieldsJSON)
	assert.False(t, model.Tags.Valid)
	assert.False(t, model.Marked)
	assert.True(t, model.CreatedAt.IsZero())
	assert.True(t, model.UpdatedAt.IsZero())
	assert.False(t, model.DeletedAt.Valid)
}

