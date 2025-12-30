package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestDeletionLogModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.DeletionLogModel{
		ID:         1,
		UserID:     100,
		ObjectType: "note",
		ObjectID:   123,
		ObjectData: `{"id":123}`,
		DeletedAt:  now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "note", model.ObjectType)
}

