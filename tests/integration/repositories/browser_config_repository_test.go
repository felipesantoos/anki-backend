package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	browserconfig "github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestBrowserConfigRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	browserConfigRepo := repositories.NewBrowserConfigRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "browser_config_save")

	sortColumn := "due"
	browserConfigEntity, err := browserconfig.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithVisibleColumns([]string{"id", "front", "back"}).
		WithColumnWidths(`{"id":100,"front":200}`).
		WithSortColumn(&sortColumn).
		WithSortDirection("asc").
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = browserConfigRepo.Save(ctx, userID, browserConfigEntity)
	require.NoError(t, err)
	assert.Greater(t, browserConfigEntity.GetID(), int64(0))
}

func TestBrowserConfigRepository_FindByUserID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	browserConfigRepo := repositories.NewBrowserConfigRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "browser_config_find")

	sortColumn := "ease"
	browserConfigEntity, err := browserconfig.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithVisibleColumns([]string{"id", "due"}).
		WithColumnWidths(`{"id":150}`).
		WithSortColumn(&sortColumn).
		WithSortDirection("desc").
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = browserConfigRepo.Save(ctx, userID, browserConfigEntity)
	require.NoError(t, err)

	found, err := browserConfigRepo.FindByUserID(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userID, found.GetUserID())
}

