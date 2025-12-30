package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestReviewRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)
	reviewRepo := repositories.NewReviewRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "review_save")

	deckID, err := deckRepo.CreateDefaultDeck(ctx, userID)
	require.NoError(t, err)

	noteType, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Template 1"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteTypeRepo.Save(ctx, userID, noteType)
	require.NoError(t, err)

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440020")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"Test"}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)

	cardEntity, err := card.NewBuilder().
		WithID(0).
		WithNoteID(noteEntity.GetID()).
		WithCardTypeID(1).
		WithDeckID(deckID).
		WithDue(time.Now().Unix() * 1000).
		WithInterval(86400).
		WithEase(2500).
		WithLapses(0).
		WithReps(0).
		WithState(valueobjects.CardStateNew).
		WithPosition(0).
		WithFlag(0).
		WithSuspended(false).
		WithBuried(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = cardRepo.Save(ctx, userID, cardEntity)
	require.NoError(t, err)

	// Create review
	reviewEntity, err := review.NewBuilder().
		WithID(0).
		WithCardID(cardEntity.GetID()).
		WithRating(3).
		WithInterval(86400).
		WithEase(2500).
		WithTimeMs(5000).
		WithType(valueobjects.ReviewTypeReview).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = reviewRepo.Save(ctx, userID, reviewEntity)
	require.NoError(t, err)
	assert.Greater(t, reviewEntity.GetID(), int64(0))
}

func TestReviewRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)
	reviewRepo := repositories.NewReviewRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "review_find")

	deckID, err := deckRepo.CreateDefaultDeck(ctx, userID)
	require.NoError(t, err)

	noteType, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Template 1"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteTypeRepo.Save(ctx, userID, noteType)
	require.NoError(t, err)

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440021")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"Test"}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)

	cardEntity, err := card.NewBuilder().
		WithID(0).
		WithNoteID(noteEntity.GetID()).
		WithCardTypeID(1).
		WithDeckID(deckID).
		WithDue(time.Now().Unix() * 1000).
		WithInterval(86400).
		WithEase(2500).
		WithLapses(0).
		WithReps(0).
		WithState(valueobjects.CardStateNew).
		WithPosition(0).
		WithFlag(0).
		WithSuspended(false).
		WithBuried(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = cardRepo.Save(ctx, userID, cardEntity)
	require.NoError(t, err)

	reviewEntity, err := review.NewBuilder().
		WithID(0).
		WithCardID(cardEntity.GetID()).
		WithRating(3).
		WithInterval(86400).
		WithEase(2500).
		WithTimeMs(5000).
		WithType(valueobjects.ReviewTypeReview).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = reviewRepo.Save(ctx, userID, reviewEntity)
	require.NoError(t, err)
	reviewID := reviewEntity.GetID()

	// Find by ID
	found, err := reviewRepo.FindByID(ctx, userID, reviewID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, reviewID, found.GetID())
	assert.Equal(t, cardEntity.GetID(), found.GetCardID())
}

func TestReviewRepository_FindByCardID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)
	reviewRepo := repositories.NewReviewRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "review_cardid")

	deckID, err := deckRepo.CreateDefaultDeck(ctx, userID)
	require.NoError(t, err)

	noteType, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Template 1"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteTypeRepo.Save(ctx, userID, noteType)
	require.NoError(t, err)

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440022")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"Test"}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)

	cardEntity, err := card.NewBuilder().
		WithID(0).
		WithNoteID(noteEntity.GetID()).
		WithCardTypeID(1).
		WithDeckID(deckID).
		WithDue(time.Now().Unix() * 1000).
		WithInterval(86400).
		WithEase(2500).
		WithLapses(0).
		WithReps(0).
		WithState(valueobjects.CardStateNew).
		WithPosition(0).
		WithFlag(0).
		WithSuspended(false).
		WithBuried(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = cardRepo.Save(ctx, userID, cardEntity)
	require.NoError(t, err)

	// Create two reviews for the same card
	review1, err := review.NewBuilder().
		WithID(0).
		WithCardID(cardEntity.GetID()).
		WithRating(3).
		WithInterval(86400).
		WithEase(2500).
		WithTimeMs(5000).
		WithType(valueobjects.ReviewTypeReview).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = reviewRepo.Save(ctx, userID, review1)
	require.NoError(t, err)

	review2, err := review.NewBuilder().
		WithID(0).
		WithCardID(cardEntity.GetID()).
		WithRating(4).
		WithInterval(172800).
		WithEase(2600).
		WithTimeMs(6000).
		WithType(valueobjects.ReviewTypeReview).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = reviewRepo.Save(ctx, userID, review2)
	require.NoError(t, err)

	// Find by card ID
	reviews, err := reviewRepo.FindByCardID(ctx, userID, cardEntity.GetID())
	require.NoError(t, err)
	assert.Len(t, reviews, 2)
}

