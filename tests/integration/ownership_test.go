package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestOwnership_DeckRepository_Isolation(t *testing.T) {
	if os.Getenv("SKIP_DB_TESTS") == "true" {
		t.Skip("Skipping database integration tests")
	}

	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)
	ctx := context.Background()

	// Create two users first
	user1Email := "test_ownership_user1@example.com"
	user2Email := "test_ownership_user2@example.com"
	
	// Create user 1
	email1, err := valueobjects.NewEmail(user1Email)
	require.NoError(t, err)
	password1, err := valueobjects.NewPassword("password123")
	require.NoError(t, err)
	user1 := &entities.User{
		Email:         email1,
		PasswordHash:  password1,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = userRepo.Save(ctx, user1)
	require.NoError(t, err, "Failed to create user 1")
	user1ID := user1.ID

	// Create user 2
	email2, err := valueobjects.NewEmail(user2Email)
	require.NoError(t, err)
	password2, err := valueobjects.NewPassword("password123")
	require.NoError(t, err)
	user2 := &entities.User{
		Email:         email2,
		PasswordHash:  password2,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = userRepo.Save(ctx, user2)
	require.NoError(t, err, "Failed to create user 2")
	user2ID := user2.ID

	// Create decks for user 1
	deck1ID, err := deckRepo.CreateDefaultDeck(ctx, user1ID)
	require.NoError(t, err, "Failed to create deck for user 1")

	deck2ID, err := deckRepo.CreateDefaultDeck(ctx, user1ID)
	require.NoError(t, err, "Failed to create second deck for user 1")

	// Create deck for user 2
	deck3ID, err := deckRepo.CreateDefaultDeck(ctx, user2ID)
	require.NoError(t, err, "Failed to create deck for user 2")

	t.Run("user can access own decks", func(t *testing.T) {
		// User 1 should be able to access their own decks
		deck1, err := deckRepo.FindByID(ctx, user1ID, deck1ID)
		require.NoError(t, err, "User 1 should be able to access their own deck")
		assert.NotNil(t, deck1)
		assert.Equal(t, user1ID, deck1.UserID)

		deck2, err := deckRepo.FindByID(ctx, user1ID, deck2ID)
		require.NoError(t, err, "User 1 should be able to access their second deck")
		assert.NotNil(t, deck2)
		assert.Equal(t, user1ID, deck2.UserID)
	})

	t.Run("user cannot access other user's decks", func(t *testing.T) {
		// User 1 should NOT be able to access user 2's deck
		deck, err := deckRepo.FindByID(ctx, user1ID, deck3ID)
		assert.Error(t, err, "User 1 should not be able to access user 2's deck")
		assert.Nil(t, deck)
		assert.ErrorIs(t, err, ownership.ErrResourceNotFound, "Should return ErrResourceNotFound")

		// User 2 should NOT be able to access user 1's decks
		deck, err = deckRepo.FindByID(ctx, user2ID, deck1ID)
		assert.Error(t, err, "User 2 should not be able to access user 1's deck")
		assert.Nil(t, deck)
		assert.ErrorIs(t, err, ownership.ErrResourceNotFound, "Should return ErrResourceNotFound")
	})

	t.Run("FindByUserID returns only user's decks", func(t *testing.T) {
		// User 1 should only see their own decks
		user1Decks, err := deckRepo.FindByUserID(ctx, user1ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(user1Decks), 2, "User 1 should have at least 2 decks")
		for _, deck := range user1Decks {
			assert.Equal(t, user1ID, deck.UserID, "All decks should belong to user 1")
		}

		// User 2 should only see their own decks
		user2Decks, err := deckRepo.FindByUserID(ctx, user2ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(user2Decks), 1, "User 2 should have at least 1 deck")
		for _, deck := range user2Decks {
			assert.Equal(t, user2ID, deck.UserID, "All decks should belong to user 2")
		}

		// Verify no overlap
		user1DeckIDs := make(map[int64]bool)
		for _, deck := range user1Decks {
			user1DeckIDs[deck.ID] = true
		}
		for _, deck := range user2Decks {
			assert.False(t, user1DeckIDs[deck.ID], "User 2's decks should not appear in user 1's list")
		}
	})

	t.Run("user cannot update other user's deck", func(t *testing.T) {
		// Try to update user 2's deck as user 1
		deck := &secondary.DeckData{
			ID:          deck3ID,
			UserID:      user2ID, // This is user 2's deck
			Name:        "Hacked Deck",
			OptionsJSON: "{}",
		}

		err := deckRepo.Save(ctx, user1ID, deck) // user1ID trying to update
		assert.Error(t, err, "User 1 should not be able to update user 2's deck")
		assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	})

	t.Run("user cannot delete other user's deck", func(t *testing.T) {
		// Try to delete user 2's deck as user 1
		err := deckRepo.Delete(ctx, user1ID, deck3ID)
		assert.Error(t, err, "User 1 should not be able to delete user 2's deck")
		assert.ErrorIs(t, err, ownership.ErrResourceNotFound)

		// Verify deck still exists for user 2
		deck, err := deckRepo.FindByID(ctx, user2ID, deck3ID)
		require.NoError(t, err, "User 2's deck should still exist")
		assert.NotNil(t, deck)
	})
}

