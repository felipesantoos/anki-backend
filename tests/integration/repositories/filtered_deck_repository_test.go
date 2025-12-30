package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	filtereddeck "github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestFilteredDeckRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	filteredDeckRepo := repositories.NewFilteredDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "filtered_deck_save")

	filteredDeckEntity, err := filtereddeck.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Due Cards").
		WithSearchFilter("is:due").
		WithLimitCards(20).
		WithOrderBy("due").
		WithReschedule(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = filteredDeckRepo.Save(ctx, userID, filteredDeckEntity)
	require.NoError(t, err)
	assert.Greater(t, filteredDeckEntity.GetID(), int64(0))
}

func TestFilteredDeckRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	filteredDeckRepo := repositories.NewFilteredDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "filtered_deck_find")

	filteredDeckEntity, err := filtereddeck.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Test Filtered Deck").
		WithSearchFilter("deck:current").
		WithLimitCards(10).
		WithOrderBy("ease").
		WithReschedule(true).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = filteredDeckRepo.Save(ctx, userID, filteredDeckEntity)
	require.NoError(t, err)
	filteredDeckID := filteredDeckEntity.GetID()

	found, err := filteredDeckRepo.FindByID(ctx, userID, filteredDeckID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, filteredDeckID, found.GetID())
}

