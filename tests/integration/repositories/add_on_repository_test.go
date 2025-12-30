package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	addon "github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestAddOnRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	addOnRepo := repositories.NewAddOnRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "addon_save")

	addOnEntity, err := addon.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithCode("1234567890").
		WithName("Test Add-on").
		WithVersion("1.0.0").
		WithEnabled(true).
		WithConfigJSON(`{"setting":"value"}`).
		WithInstalledAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = addOnRepo.Save(ctx, userID, addOnEntity)
	require.NoError(t, err)
	assert.Greater(t, addOnEntity.GetID(), int64(0))
}

func TestAddOnRepository_FindByCode(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	addOnRepo := repositories.NewAddOnRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "addon_code")

	addOnEntity, err := addon.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithCode("unique_code_123").
		WithName("Unique Add-on").
		WithVersion("2.0.0").
		WithEnabled(true).
		WithConfigJSON(`{}`).
		WithInstalledAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = addOnRepo.Save(ctx, userID, addOnEntity)
	require.NoError(t, err)

	found, err := addOnRepo.FindByCode(ctx, userID, "unique_code_123")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "unique_code_123", found.GetCode())
}

func TestAddOnRepository_FindEnabled(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	addOnRepo := repositories.NewAddOnRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "addon_enabled")

	// Create enabled add-on
	addOn1, err := addon.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithCode("enabled_1").
		WithName("Enabled Add-on").
		WithVersion("1.0.0").
		WithEnabled(true).
		WithConfigJSON(`{}`).
		WithInstalledAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = addOnRepo.Save(ctx, userID, addOn1)
	require.NoError(t, err)

	// Create disabled add-on
	addOn2, err := addon.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithCode("disabled_1").
		WithName("Disabled Add-on").
		WithVersion("1.0.0").
		WithEnabled(false).
		WithConfigJSON(`{}`).
		WithInstalledAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = addOnRepo.Save(ctx, userID, addOn2)
	require.NoError(t, err)

	// Find enabled
	enabled, err := addOnRepo.FindEnabled(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, enabled, 1)
	assert.True(t, enabled[0].GetEnabled())
}

