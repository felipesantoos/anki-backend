package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestMediaRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	mediaRepo := repositories.NewMediaRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "media_save")

	mediaEntity, err := media.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("image.jpg").
		WithHash("abc123def456").
		WithSize(1024000).
		WithMimeType("image/jpeg").
		WithStoragePath("/storage/image.jpg").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = mediaRepo.Save(ctx, userID, mediaEntity)
	require.NoError(t, err)
	assert.Greater(t, mediaEntity.GetID(), int64(0))
}

func TestMediaRepository_FindByHash(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	mediaRepo := repositories.NewMediaRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "media_hash")

	mediaEntity, err := media.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("test.jpg").
		WithHash("unique_hash_123").
		WithSize(500000).
		WithMimeType("image/jpeg").
		WithStoragePath("/storage/test.jpg").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = mediaRepo.Save(ctx, userID, mediaEntity)
	require.NoError(t, err)

	found, err := mediaRepo.FindByHash(ctx, userID, "unique_hash_123")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "unique_hash_123", found.GetHash())
}

func TestMediaRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	mediaRepo := repositories.NewMediaRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "media_delete")

	mediaEntity, err := media.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("delete.jpg").
		WithHash("delete_hash").
		WithSize(100000).
		WithMimeType("image/jpeg").
		WithStoragePath("/storage/delete.jpg").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = mediaRepo.Save(ctx, userID, mediaEntity)
	require.NoError(t, err)
	mediaID := mediaEntity.GetID()

	err = mediaRepo.Delete(ctx, userID, mediaID)
	require.NoError(t, err)

	found, err := mediaRepo.FindByID(ctx, userID, mediaID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)
}

