package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMediaModel_Creation(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	model := &MediaModel{
		ID:          1,
		UserID:      100,
		Filename:    "image.jpg",
		Hash:        "abc123",
		Size:        1024000,
		MimeType:    "image/jpeg",
		StoragePath: "/storage/image.jpg",
		CreatedAt:   now,
		DeletedAt:   sqlNullTime(deletedAt, true),
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "image.jpg", model.Filename)
	assert.True(t, model.DeletedAt.Valid)
}

func TestMediaModel_NullDeletedAt(t *testing.T) {
	model := &MediaModel{
		ID:    2,
		UserID: 200,
		DeletedAt: sqlNullTime(time.Time{}, false),
	}

	assert.False(t, model.DeletedAt.Valid)
}
