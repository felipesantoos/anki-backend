package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deckoptionspreset "github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestDeckOptionsPresetRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	presetRepo := repositories.NewDeckOptionsPresetRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "preset_save")

	presetEntity, err := deckoptionspreset.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Default Preset").
		WithOptionsJSON(`{"newCardsPerDay":20}`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = presetRepo.Save(ctx, userID, presetEntity)
	require.NoError(t, err)
	assert.Greater(t, presetEntity.GetID(), int64(0))
}

func TestDeckOptionsPresetRepository_FindByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	presetRepo := repositories.NewDeckOptionsPresetRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "preset_name")

	presetEntity, err := deckoptionspreset.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Unique Preset").
		WithOptionsJSON(`{"newCardsPerDay":10}`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = presetRepo.Save(ctx, userID, presetEntity)
	require.NoError(t, err)

	found, err := presetRepo.FindByName(ctx, userID, "Unique Preset")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Unique Preset", found.GetName())
}

