package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestCheckDatabaseLogModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.CheckDatabaseLogModel{
		ID:            1,
		UserID:        100,
		Status:         "completed",
		IssuesFound:    0,
		IssuesDetails:  `[]`,
		CreatedAt:     now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "completed", model.Status)
}

