package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestProfileRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	profileRepo := repositories.NewProfileRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "profile_save")

	profileEntity, err := profile.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Default Profile").
		WithAnkiWebSyncEnabled(true).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = profileRepo.Save(ctx, userID, profileEntity)
	require.NoError(t, err)
	assert.Greater(t, profileEntity.GetID(), int64(0))
}

func TestProfileRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	profileRepo := repositories.NewProfileRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "profile_find")

	profileEntity, err := profile.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Test Profile").
		WithAnkiWebSyncEnabled(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = profileRepo.Save(ctx, userID, profileEntity)
	require.NoError(t, err)
	profileID := profileEntity.GetID()

	found, err := profileRepo.FindByID(ctx, userID, profileID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, profileID, found.GetID())
	assert.Equal(t, "Test Profile", found.GetName())
}

func TestProfileRepository_FindByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	profileRepo := repositories.NewProfileRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "profile_name")

	profileEntity, err := profile.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Unique Profile").
		WithAnkiWebSyncEnabled(true).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = profileRepo.Save(ctx, userID, profileEntity)
	require.NoError(t, err)

	found, err := profileRepo.FindByName(ctx, userID, "Unique Profile")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Unique Profile", found.GetName())
}

func TestProfileRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	profileRepo := repositories.NewProfileRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "profile_update")

	profileEntity, err := profile.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Original Name").
		WithAnkiWebSyncEnabled(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = profileRepo.Save(ctx, userID, profileEntity)
	require.NoError(t, err)
	profileID := profileEntity.GetID()

	profileEntity.SetName("Updated Name")
	profileEntity.SetAnkiWebSyncEnabled(true)
	err = profileRepo.Update(ctx, userID, profileID, profileEntity)
	require.NoError(t, err)

	updated, err := profileRepo.FindByID(ctx, userID, profileID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.GetName())
	assert.True(t, updated.GetAnkiWebSyncEnabled())
}

func TestProfileRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	profileRepo := repositories.NewProfileRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "profile_delete")

	profileEntity, err := profile.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("To Delete").
		WithAnkiWebSyncEnabled(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = profileRepo.Save(ctx, userID, profileEntity)
	require.NoError(t, err)
	profileID := profileEntity.GetID()

	err = profileRepo.Delete(ctx, userID, profileID)
	require.NoError(t, err)

	found, err := profileRepo.FindByID(ctx, userID, profileID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)
}

