package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestBackupModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.BackupModel{
		ID:          1,
		UserID:      100,
		Filename:    "backup.apkg",
		Size:        5000000,
		StoragePath: "/storage/backup.apkg",
		BackupType:  "automatic",
		CreatedAt:   now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "backup.apkg", model.Filename)
}

