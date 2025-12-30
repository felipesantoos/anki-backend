package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestProfileModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.ProfileModel{
		ID:                 1,
		UserID:             100,
		Name:               "Default Profile",
		AnkiWebSyncEnabled: true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Default Profile", model.Name)
}

