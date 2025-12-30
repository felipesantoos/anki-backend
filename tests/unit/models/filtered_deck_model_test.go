package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestFilteredDeckModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.FilteredDeckModel{
		ID:           1,
		UserID:       100,
		Name:         "Due Cards",
		SearchFilter: "is:due",
		LimitCards:   20,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Due Cards", model.Name)
}

