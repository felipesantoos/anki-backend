package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestSharedDeckRatingRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	sharedDeckRepo := repositories.NewSharedDeckRepository(db.DB)
	ratingRepo := repositories.NewSharedDeckRatingRepository(db.DB)

	authorID, _ := createTestUser(t, ctx, userRepo, "rating_author")
	userID, _ := createTestUser(t, ctx, userRepo, "rating_user")

	// Create shared deck
	sharedDeckEntity, err := shareddeck.NewBuilder().
		WithID(0).
		WithAuthorID(authorID).
		WithName("Test Deck").
		WithPackagePath("/packages/test.apkg").
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

	// Create rating
	ratingVO, err := valueobjects.NewSharedDeckRating(5)
	require.NoError(t, err)
	comment := "Great deck!"
	ratingEntity, err := shareddeckrating.NewBuilder().
		WithID(0).
		WithSharedDeckID(sharedDeckEntity.GetID()).
		WithUserID(userID).
		WithRating(ratingVO).
		WithComment(&comment).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = ratingRepo.Save(ctx, userID, ratingEntity)
	require.NoError(t, err)
	assert.Greater(t, ratingEntity.GetID(), int64(0))
}

func TestSharedDeckRatingRepository_FindBySharedDeckID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	sharedDeckRepo := repositories.NewSharedDeckRepository(db.DB)
	ratingRepo := repositories.NewSharedDeckRatingRepository(db.DB)

	authorID, _ := createTestUser(t, ctx, userRepo, "rating_deck_author")
	userID, _ := createTestUser(t, ctx, userRepo, "rating_deck_user")

	// Create shared deck
	sharedDeckEntity, err := shareddeck.NewBuilder().
		WithID(0).
		WithAuthorID(authorID).
		WithName("Rated Deck").
		WithPackagePath("/packages/rated.apkg").
		WithPackageSize(2000000).
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

	// Create rating
	ratingVO, err := valueobjects.NewSharedDeckRating(4)
	require.NoError(t, err)
	ratingEntity, err := shareddeckrating.NewBuilder().
		WithID(0).
		WithSharedDeckID(sharedDeckEntity.GetID()).
		WithUserID(userID).
		WithRating(ratingVO).
		WithComment(nil).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = ratingRepo.Save(ctx, userID, ratingEntity)
	require.NoError(t, err)

	// Find by shared deck ID
	ratings, err := ratingRepo.FindBySharedDeckID(ctx, sharedDeckEntity.GetID(), 0, 10)
	require.NoError(t, err)
	assert.Greater(t, len(ratings), 0)
}

