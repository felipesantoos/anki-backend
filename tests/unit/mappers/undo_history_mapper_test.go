package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestUndoHistoryToDomain_WithAllFields(t *testing.T) {
	now := time.Now()

	model := &models.UndoHistoryModel{
		ID:            1,
		UserID:        100,
		OperationType: "delete_note",
		OperationData: `{"note_id":123}`,
		CreatedAt:     now,
	}

	entity, err := UndoHistoryToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "delete_note", entity.GetOperationType())
	assert.Equal(t, `{"note_id":123}`, entity.GetOperationData())
	assert.Equal(t, now, entity.GetCreatedAt())
}

func TestUndoHistoryToDomain_NilInput(t *testing.T) {
	entity, err := UndoHistoryToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestUndoHistoryToModel_WithAllFields(t *testing.T) {
	now := time.Now()

	entity, err := undohistory.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithOperationType("delete_note").
		WithOperationData(`{"note_id":123}`).
		WithCreatedAt(now).
		Build()
	require.NoError(t, err)

	model := UndoHistoryToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "delete_note", model.OperationType)
}

