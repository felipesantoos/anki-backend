package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	savedsearch "github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestSavedSearchRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	savedSearchRepo := repositories.NewSavedSearchRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "saved_search_save")

	savedSearchEntity, err := savedsearch.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Due Cards").
		WithSearchQuery("is:due").
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = savedSearchRepo.Save(ctx, userID, savedSearchEntity)
	require.NoError(t, err)
	assert.Greater(t, savedSearchEntity.GetID(), int64(0))
}

func TestSavedSearchRepository_FindByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	savedSearchRepo := repositories.NewSavedSearchRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "saved_search_name")

	savedSearchEntity, err := savedsearch.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Unique Search").
		WithSearchQuery("deck:current").
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = savedSearchRepo.Save(ctx, userID, savedSearchEntity)
	require.NoError(t, err)

	found, err := savedSearchRepo.FindByName(ctx, userID, "Unique Search")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Unique Search", found.GetName())
}

