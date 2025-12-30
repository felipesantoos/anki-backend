package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestToNoteResponse(t *testing.T) {
	now := time.Now()
	guid, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	
	n, _ := note.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithGUID(guid).
		WithNoteTypeID(1).
		WithFieldsJSON(`{"Front": "cat"}`).
		WithTags([]string{"animal"}).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToNoteResponse(n)
		assert.NotNil(t, res)
		assert.Equal(t, n.GetID(), res.ID)
		assert.Equal(t, n.GetGUID().Value(), res.GUID)
		assert.Equal(t, n.GetNoteTypeID(), res.NoteTypeID)
		assert.Equal(t, n.GetFieldsJSON(), res.FieldsJSON)
		assert.Equal(t, n.GetTags(), res.Tags)
		assert.Equal(t, n.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, n.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToNoteResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToNoteResponseList(t *testing.T) {
	guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
	guid2, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440002")
	n1, _ := note.NewBuilder().WithID(1).WithUserID(1).WithGUID(guid1).WithNoteTypeID(1).Build()
	n2, _ := note.NewBuilder().WithID(2).WithUserID(1).WithGUID(guid2).WithNoteTypeID(1).Build()
	notes := []*note.Note{n1, n2}

	res := ToNoteResponseList(notes)
	assert.Len(t, res, 2)
	assert.Equal(t, n1.GetID(), res[0].ID)
	assert.Equal(t, n2.GetID(), res[1].ID)
}
