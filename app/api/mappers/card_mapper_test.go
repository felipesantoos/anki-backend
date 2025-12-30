package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestToCardResponse(t *testing.T) {
	now := time.Now()
	
	c := &card.Card{}
	c.SetID(10)
	c.SetNoteID(1)
	c.SetDeckID(2)
	c.SetState(valueobjects.CardStateReview)
	c.SetInterval(5)
	c.SetEase(2500)
	c.SetReps(10)
	c.SetLapses(1)
	c.SetDue(now.Unix())
	c.SetPosition(0)
	c.SetFlag(1)
	c.SetSuspended(true)
	c.SetCreatedAt(now)
	c.SetUpdatedAt(now)

	t.Run("Success", func(t *testing.T) {
		res := ToCardResponse(c)
		assert.NotNil(t, res)
		assert.Equal(t, c.GetID(), res.ID)
		assert.Equal(t, c.GetNoteID(), res.NoteID)
		assert.Equal(t, c.GetDeckID(), res.DeckID)
		assert.Equal(t, string(c.GetState()), res.State)
		assert.Equal(t, c.GetInterval(), res.Interval)
		assert.Equal(t, c.GetEase(), res.Ease)
		assert.Equal(t, c.GetReps(), res.Reviews)
		assert.Equal(t, c.GetLapses(), res.Lapses)
		assert.Equal(t, c.GetDue(), res.Due)
		assert.Equal(t, c.GetPosition(), res.Ord)
		assert.Equal(t, c.GetFlag(), res.Flags)
		assert.Equal(t, c.GetSuspended(), res.Suspended)
		assert.Equal(t, c.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, c.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToCardResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToCardResponseList(t *testing.T) {
	c1 := &card.Card{}
	c1.SetID(1)
	c2 := &card.Card{}
	c2.SetID(2)
	cards := []*card.Card{c1, c2}

	res := ToCardResponseList(cards)
	assert.Len(t, res, 2)
	assert.Equal(t, c1.GetID(), res[0].ID)
	assert.Equal(t, c2.GetID(), res[1].ID)
}
