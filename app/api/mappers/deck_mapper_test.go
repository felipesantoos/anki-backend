package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/stretchr/testify/assert"
)

func TestToDeckResponse(t *testing.T) {
	now := time.Now()
	parentID := int64(1)
	
	d, _ := deck.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithName("Test Deck").
		WithParentID(&parentID).
		WithOptionsJSON(`{"key": "value"}`).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToDeckResponse(d)
		assert.NotNil(t, res)
		assert.Equal(t, d.GetID(), res.ID)
		assert.Equal(t, d.GetUserID(), res.UserID)
		assert.Equal(t, d.GetName(), res.Name)
		assert.Equal(t, d.GetParentID(), res.ParentID)
		assert.Equal(t, d.GetOptionsJSON(), res.OptionsJSON)
		assert.Equal(t, d.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, d.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToDeckResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToDeckResponseList(t *testing.T) {
	now := time.Now()
	d1, _ := deck.NewBuilder().WithID(1).WithUserID(1).WithName("D1").WithCreatedAt(now).Build()
	d2, _ := deck.NewBuilder().WithID(2).WithUserID(1).WithName("D2").WithCreatedAt(now).Build()
	decks := []*deck.Deck{d1, d2}

	res := ToDeckResponseList(decks)
	assert.Len(t, res, 2)
	assert.Equal(t, d1.GetID(), res[0].ID)
	assert.Equal(t, d2.GetID(), res[1].ID)
}
