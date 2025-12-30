package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestMediaToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	model := &models.MediaModel{
		ID:          1,
		UserID:      100,
		Filename:    "image.jpg",
		Hash:        "abc123def456",
		Size:        1024000,
		MimeType:    "image/jpeg",
		StoragePath: "/storage/media/abc123def456.jpg",
		CreatedAt:   now,
		DeletedAt:   sqlNullTime(deletedAt, true),
	}

	entity, err := MediaToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "image.jpg", entity.GetFilename())
	assert.Equal(t, "abc123def456", entity.GetHash())
	assert.Equal(t, int64(1024000), entity.GetSize())
	assert.Equal(t, "image/jpeg", entity.GetMimeType())
	assert.Equal(t, "/storage/media/abc123def456.jpg", entity.GetStoragePath())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.NotNil(t, entity.GetDeletedAt())
	assert.Equal(t, deletedAt, *entity.GetDeletedAt())
}

func TestMediaToDomain_WithNullDeletedAt(t *testing.T) {
	now := time.Now()

	model := &models.MediaModel{
		ID:          2,
		UserID:      200,
		Filename:    "video.mp4",
		Hash:        "xyz789",
		Size:        5000000,
		MimeType:    "video/mp4",
		StoragePath: "/storage/media/xyz789.mp4",
		CreatedAt:   now,
		DeletedAt:   sqlNullTime(time.Time{}, false),
	}

	entity, err := MediaToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetDeletedAt())
}

func TestMediaToDomain_NilInput(t *testing.T) {
	entity, err := MediaToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestMediaToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	entity, err := media.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithFilename("image.jpg").
		WithHash("abc123def456").
		WithSize(1024000).
		WithMimeType("image/jpeg").
		WithStoragePath("/storage/media/abc123def456.jpg").
		WithCreatedAt(now).
		WithDeletedAt(&deletedAt).
		Build()
	require.NoError(t, err)

	model := MediaToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "image.jpg", model.Filename)
	assert.Equal(t, "abc123def456", model.Hash)
	assert.Equal(t, int64(1024000), model.Size)
	assert.Equal(t, "image/jpeg", model.MimeType)
	assert.Equal(t, "/storage/media/abc123def456.jpg", model.StoragePath)
	assert.Equal(t, now, model.CreatedAt)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestMediaToModel_WithNullDeletedAt(t *testing.T) {
	now := time.Now()

	entity, err := media.NewBuilder().
		WithID(2).
		WithUserID(200).
		WithFilename("video.mp4").
		WithHash("xyz789").
		WithSize(5000000).
		WithMimeType("video/mp4").
		WithStoragePath("/storage/media/xyz789.mp4").
		WithCreatedAt(now).
		WithDeletedAt(nil).
		Build()
	require.NoError(t, err)

	model := MediaToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.DeletedAt.Valid)
}

