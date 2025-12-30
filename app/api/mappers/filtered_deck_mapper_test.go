package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/stretchr/testify/assert"
)

func TestToFilteredDeckResponse(t *testing.T) {
	now := time.Now()
	
	fd, _ := filtereddeck.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithName("Filtered").
		WithSearchFilter("is:due").
		WithLimitCards(100).
		WithOrderBy("random").
		WithReschedule(true).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToFilteredDeckResponse(fd)
		assert.NotNil(t, res)
		assert.Equal(t, fd.GetID(), res.ID)
		assert.Equal(t, fd.GetUserID(), res.UserID)
		assert.Equal(t, fd.GetName(), res.Name)
		assert.Equal(t, fd.GetSearchFilter(), res.SearchFilter)
		assert.Equal(t, fd.GetLimitCards(), res.Limit)
		assert.Equal(t, fd.GetOrderBy(), res.OrderBy)
		assert.Equal(t, fd.GetReschedule(), res.Reschedule)
		assert.Equal(t, fd.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, fd.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToFilteredDeckResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToFilteredDeckResponseList(t *testing.T) {
	fd1, _ := filtereddeck.NewBuilder().WithID(1).WithUserID(1).WithName("FD1").Build()
	fd2, _ := filtereddeck.NewBuilder().WithID(2).WithUserID(1).WithName("FD2").Build()
	decks := []*filtereddeck.FilteredDeck{fd1, fd2}

	res := ToFilteredDeckResponseList(decks)
	assert.Len(t, res, 2)
	assert.Equal(t, fd1.GetID(), res[0].ID)
	assert.Equal(t, fd2.GetID(), res[1].ID)
}
