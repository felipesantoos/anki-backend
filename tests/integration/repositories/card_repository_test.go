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
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestCardRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "card_save")

	// Create deck
	deckID, err := deckRepo.CreateDefaultDeck(ctx, userID)
	require.NoError(t, err)

	// Create note type
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

	// Create note
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440010")
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

	// Create card
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
	assert.Greater(t, cardEntity.GetID(), int64(0))
}

func TestCardRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "card_find")

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

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440011")
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
	cardID := cardEntity.GetID()

	// Find by ID
	found, err := cardRepo.FindByID(ctx, userID, cardID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, cardID, found.GetID())
}

func TestCardRepository_FindByNoteID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "card_noteid")

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

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440012")
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

	// Create two cards for the same note
	card1, err := card.NewBuilder().
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
	err = cardRepo.Save(ctx, userID, card1)
	require.NoError(t, err)

	card2, err := card.NewBuilder().
		WithID(0).
		WithNoteID(noteEntity.GetID()).
		WithCardTypeID(2).
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
	err = cardRepo.Save(ctx, userID, card2)
	require.NoError(t, err)

	// Find by note ID
	cards, err := cardRepo.FindByNoteID(ctx, userID, noteEntity.GetID())
	require.NoError(t, err)
	assert.Len(t, cards, 2)
}

func TestCardRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "card_update")

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

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440013")
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
	cardID := cardEntity.GetID()

	// Update card
	cardEntity.SetEase(2600)
	cardEntity.SetReps(1)
	cardEntity.SetState(valueobjects.CardStateReview)
	err = cardRepo.Update(ctx, userID, cardID, cardEntity)
	require.NoError(t, err)

	// Verify update
	updated, err := cardRepo.FindByID(ctx, userID, cardID)
	require.NoError(t, err)
	assert.Equal(t, 2600, updated.GetEase())
	assert.Equal(t, 1, updated.GetReps())
	assert.Equal(t, valueobjects.CardStateReview, updated.GetState())
}

func TestCardRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "card_delete")

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

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440014")
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
	cardID := cardEntity.GetID()

	// Delete
	err = cardRepo.Delete(ctx, userID, cardID)
	require.NoError(t, err)

	// Verify deletion
	found, err := cardRepo.FindByID(ctx, userID, cardID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)
}

func TestCardRepository_CountByDeckAndState(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "card_count")

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

	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440015")
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

	// Create 3 new cards and 2 review cards
	for i := 0; i < 3; i++ {
		c, _ := card.NewBuilder().
			WithID(0).
			WithNoteID(noteEntity.GetID()).
			WithCardTypeID(i + 1).
			WithDeckID(deckID).
			WithEase(2500).
			WithDue(int64(i)). // For new cards, due is position
			WithState(valueobjects.CardStateNew).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		err = cardRepo.Save(ctx, userID, c)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		c, _ := card.NewBuilder().
			WithID(0).
			WithNoteID(noteEntity.GetID()).
			WithCardTypeID(i + 10).
			WithDeckID(deckID).
			WithEase(2500).
			WithDue(time.Now().Unix() * 1000). // For review cards, due is timestamp
			WithState(valueobjects.CardStateReview).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		err = cardRepo.Save(ctx, userID, c)
		require.NoError(t, err)
	}

	// Count new cards
	newCount, err := cardRepo.CountByDeckAndState(ctx, userID, deckID, valueobjects.CardStateNew)
	require.NoError(t, err)
	assert.Equal(t, 3, newCount)

	// Count review cards
	reviewCount, err := cardRepo.CountByDeckAndState(ctx, userID, deckID, valueobjects.CardStateReview)
	require.NoError(t, err)
	assert.Equal(t, 2, reviewCount)

	// Count learning cards (should be 0)
	learnCount, err := cardRepo.CountByDeckAndState(ctx, userID, deckID, valueobjects.CardStateLearn)
	require.NoError(t, err)
	assert.Equal(t, 0, learnCount)
}

