package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/stretchr/testify/assert"
)

func TestToUndoHistoryResponse(t *testing.T) {
	now := time.Now()
	
	uh, _ := undohistory.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithOperationType(undohistory.OperationTypeEditNote).
		WithOperationData(`{"id": 1}`).
		WithCreatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToUndoHistoryResponse(uh)
		assert.NotNil(t, res)
		assert.Equal(t, uh.GetID(), res.ID)
		assert.Equal(t, uh.GetUserID(), res.UserID)
		assert.Equal(t, uh.GetOperationType(), res.OperationType)
		assert.Equal(t, uh.GetOperationData(), res.OperationData)
		assert.Equal(t, uh.GetCreatedAt(), res.CreatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToUndoHistoryResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToUndoHistoryResponseList(t *testing.T) {
	uh1, _ := undohistory.NewBuilder().WithID(1).WithUserID(1).WithOperationType(undohistory.OperationTypeEditNote).Build()
	uh2, _ := undohistory.NewBuilder().WithID(2).WithUserID(1).WithOperationType(undohistory.OperationTypeEditNote).Build()
	history := []*undohistory.UndoHistory{uh1, uh2}

	res := ToUndoHistoryResponseList(history)
	assert.Len(t, res, 2)
	assert.Equal(t, uh1.GetID(), res[0].ID)
	assert.Equal(t, uh2.GetID(), res[1].ID)
}
