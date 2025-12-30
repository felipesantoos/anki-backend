package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestSharedDeckModel_Creation(t *testing.T) {
	now := time.Now()
	model := &models.SharedDeckModel{
		ID:            1,
		AuthorID:      100,
		Name:          "Spanish Vocabulary",
		PackagePath:   "/packages/spanish.apkg",
		PackageSize:   5000000,
		DownloadCount: 100,
		RatingAverage: 4.5,
		RatingCount:   20,
		IsFeatured:    true,
		IsPublic:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Spanish Vocabulary", model.Name)
}

