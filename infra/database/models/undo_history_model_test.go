package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUndoHistoryModel_Creation(t *testing.T) {
	now := time.Now()
	model := &UndoHistoryModel{
		ID:            1,
		UserID:        100,
		OperationType: "delete_note",
		OperationData: `{"note_id":123}`,
		CreatedAt:     now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "delete_note", model.OperationType)
}

