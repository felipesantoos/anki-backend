package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestSyncMetaRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	syncMetaRepo := repositories.NewSyncMetaRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "sync_save")

	syncMetaEntity, err := syncmeta.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithClientID("client-123").
		WithLastSync(time.Now()).
		WithLastSyncUSN(0).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = syncMetaRepo.Save(ctx, userID, syncMetaEntity)
	require.NoError(t, err)
	assert.Greater(t, syncMetaEntity.GetID(), int64(0))
}

func TestSyncMetaRepository_FindByClientID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	syncMetaRepo := repositories.NewSyncMetaRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "sync_clientid")

	syncMetaEntity, err := syncmeta.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithClientID("unique_client_456").
		WithLastSync(time.Now()).
		WithLastSyncUSN(42).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = syncMetaRepo.Save(ctx, userID, syncMetaEntity)
	require.NoError(t, err)

	found, err := syncMetaRepo.FindByClientID(ctx, userID, "unique_client_456")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "unique_client_456", found.GetClientID())
	assert.Equal(t, int64(42), found.GetLastSyncUSN())
}

