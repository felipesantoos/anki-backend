package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/stretchr/testify/assert"
)

func TestToNoteTypeResponse(t *testing.T) {
	now := time.Now()
	
	nt, _ := notetype.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithName("Basic").
		WithFieldsJSON(`[]`).
		WithCardTypesJSON(`[]`).
		WithTemplatesJSON(`{}`).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToNoteTypeResponse(nt)
		assert.NotNil(t, res)
		assert.Equal(t, nt.GetID(), res.ID)
		assert.Equal(t, nt.GetUserID(), res.UserID)
		assert.Equal(t, nt.GetName(), res.Name)
		assert.Equal(t, nt.GetFieldsJSON(), res.FieldsJSON)
		assert.Equal(t, nt.GetCardTypesJSON(), res.CardTypesJSON)
		assert.Equal(t, nt.GetTemplatesJSON(), res.TemplatesJSON)
		assert.Equal(t, nt.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, nt.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToNoteTypeResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToNoteTypeResponseList(t *testing.T) {
	nt1, _ := notetype.NewBuilder().WithID(1).WithUserID(1).WithName("NT1").Build()
	nt2, _ := notetype.NewBuilder().WithID(2).WithUserID(1).WithName("NT2").Build()
	noteTypes := []*notetype.NoteType{nt1, nt2}

	res := ToNoteTypeResponseList(noteTypes)
	assert.Len(t, res, 2)
	assert.Equal(t, nt1.GetID(), res[0].ID)
	assert.Equal(t, nt2.GetID(), res[1].ID)
}
