package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestSharedDeckRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	sharedDeckRepo := repositories.NewSharedDeckRepository(db.DB)

	authorID, _ := createTestUser(t, ctx, userRepo, "shared_deck_save")

	description := "A deck for learning Spanish"
	category := "Languages"
	sharedDeckEntity, err := shareddeck.NewBuilder().
		WithID(0).
		WithAuthorID(authorID).
		WithName("Spanish Vocabulary").
		WithDescription(&description).
		WithCategory(&category).
		WithPackagePath("/packages/spanish.apkg").
		WithPackageSize(5000000).
		WithDownloadCount(0).
		WithRatingAverage(0).
		WithRatingCount(0).
		WithTags([]string{"spanish", "vocabulary"}).
		WithIsFeatured(false).
		WithIsPublic(true).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = sharedDeckRepo.Save(ctx, authorID, sharedDeckEntity)
	require.NoError(t, err)
	assert.Greater(t, sharedDeckEntity.GetID(), int64(0))
}

func TestSharedDeckRepository_FindPublic(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	sharedDeckRepo := repositories.NewSharedDeckRepository(db.DB)

	authorID, _ := createTestUser(t, ctx, userRepo, "shared_deck_public")

	sharedDeckEntity, err := shareddeck.NewBuilder().
		WithID(0).
		WithAuthorID(authorID).
		WithName("Public Deck").
		WithPackagePath("/packages/public.apkg").
		WithPackageSize(1000000).
		WithDownloadCount(0).
		WithRatingAverage(0).
		WithRatingCount(0).
		WithTags([]string{}).
		WithIsFeatured(false).
		WithIsPublic(true).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = sharedDeckRepo.Save(ctx, authorID, sharedDeckEntity)
	require.NoError(t, err)

	public, err := sharedDeckRepo.FindPublic(ctx, 10, 0)
	require.NoError(t, err)
	assert.Greater(t, len(public), 0)
}

