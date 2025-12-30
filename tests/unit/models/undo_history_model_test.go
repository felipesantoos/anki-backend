package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestUndoHistoryModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.UndoHistoryModel{
		ID:            1,
		UserID:        100,
		OperationType: "delete_note",
		OperationData: `{"note_id":123}`,
		CreatedAt:     now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "delete_note", model.OperationType)
}

