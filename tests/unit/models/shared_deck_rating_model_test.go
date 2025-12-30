package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestSharedDeckRatingModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.SharedDeckRatingModel{
		ID:          1,
		SharedDeckID: 100,
		UserID:      200,
		Rating:      5,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, 5, model.Rating)
}

