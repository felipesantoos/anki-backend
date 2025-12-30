package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flagname "github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestFlagNameRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	flagNameRepo := repositories.NewFlagNameRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "flag_name_save")

	flagNameEntity, err := flagname.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFlagNumber(1).
		WithName("Important").
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = flagNameRepo.Save(ctx, userID, flagNameEntity)
	require.NoError(t, err)
	assert.Greater(t, flagNameEntity.GetID(), int64(0))
}

func TestFlagNameRepository_FindByFlagNumber(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	flagNameRepo := repositories.NewFlagNameRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "flag_name_number")

	flagNameEntity, err := flagname.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFlagNumber(2).
		WithName("Review Later").
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = flagNameRepo.Save(ctx, userID, flagNameEntity)
	require.NoError(t, err)

	found, err := flagNameRepo.FindByFlagNumber(ctx, userID, 2)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, 2, found.GetFlagNumber())
}

