package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestSyncMetaModel_Creation(t *testing.T) {
	now := time.Now()

	model := &models.SyncMetaModel{
		ID:          1,
		UserID:      100,
		ClientID:    "client-123",
		LastSync:    now,
		LastSyncUSN: 42,
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Hour),
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "client-123", model.ClientID)
	assert.Equal(t, int64(42), model.LastSyncUSN)
}
