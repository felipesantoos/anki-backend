package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestDeckRepository_CreateDefaultDeck(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "default_deck")

	deckID, err := deckRepo.CreateDefaultDeck(ctx, userID)
	require.NoError(t, err)
	assert.Greater(t, deckID, int64(0))

	found, err := deckRepo.FindByID(ctx, userID, deckID)
	require.NoError(t, err)
	assert.Equal(t, "Default", found.GetName())
	assert.True(t, found.IsRoot())
}

func TestDeckRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_save")

	deckEntity := &deck.Deck{}
	deckEntity.SetName("Custom Deck")
	deckEntity.SetUserID(userID)
	deckEntity.SetOptionsJSON("{}")
	deckEntity.SetCreatedAt(time.Now())
	deckEntity.SetUpdatedAt(time.Now())

	err := deckRepo.Save(ctx, userID, deckEntity)
	require.NoError(t, err)
	assert.Greater(t, deckEntity.GetID(), int64(0))

	found, err := deckRepo.FindByID(ctx, userID, deckEntity.GetID())
	require.NoError(t, err)
	assert.Equal(t, "Custom Deck", found.GetName())
}

func TestDeckRepository_Hierarchy(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_hierarchy")

	// Create Parent
	parent := &deck.Deck{}
	parent.SetName("Parent")
	parent.SetUserID(userID)
	parent.SetOptionsJSON("{}")
	parent.SetCreatedAt(time.Now())
	parent.SetUpdatedAt(time.Now())
	_ = deckRepo.Save(ctx, userID, parent)

	// Create Child
	child := &deck.Deck{}
	child.SetName("Child")
	child.SetUserID(userID)
	parentID := parent.GetID()
	child.SetParentID(&parentID)
	child.SetOptionsJSON("{}")
	child.SetCreatedAt(time.Now())
	child.SetUpdatedAt(time.Now())
	_ = deckRepo.Save(ctx, userID, child)

	// Test FindByParentID
	children, err := deckRepo.FindByParentID(ctx, userID, parentID)
	require.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, child.GetID(), children[0].GetID())

	// Test GetFullPath
	allDecks, _ := deckRepo.FindByUserID(ctx, userID, "")
	assert.Equal(t, "Parent::Child", child.GetFullPath(allDecks))
}

func TestDeckRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_update")

	deckEntity := &deck.Deck{}
	deckEntity.SetName("Old Name")
	deckEntity.SetUserID(userID)
	deckEntity.SetOptionsJSON("{}")
	_ = deckRepo.Save(ctx, userID, deckEntity)

	deckEntity.SetName("New Name")
	err := deckRepo.Update(ctx, userID, deckEntity.GetID(), deckEntity)
	require.NoError(t, err)

	found, _ := deckRepo.FindByID(ctx, userID, deckEntity.GetID())
	assert.Equal(t, "New Name", found.GetName())
}

func TestDeckRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_delete")

	deckEntity := &deck.Deck{}
	deckEntity.SetName("To Delete")
	deckEntity.SetUserID(userID)
	deckEntity.SetOptionsJSON("{}")
	_ = deckRepo.Save(ctx, userID, deckEntity)

	err := deckRepo.Delete(ctx, userID, deckEntity.GetID())
	require.NoError(t, err)

	found, err := deckRepo.FindByID(ctx, userID, deckEntity.GetID())
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found) // Soft deleted
}

func TestDeckRepository_Delete_Recursive(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_delete_recursive")

	// Create Hierarchy: Parent -> Child -> Grandchild
	parent := &deck.Deck{}
	parent.SetName("Parent")
	parent.SetUserID(userID)
	parent.SetOptionsJSON("{}")
	_ = deckRepo.Save(ctx, userID, parent)

	parentID := parent.GetID()
	child := &deck.Deck{}
	child.SetName("Child")
	child.SetUserID(userID)
	child.SetParentID(&parentID)
	child.SetOptionsJSON("{}")
	_ = deckRepo.Save(ctx, userID, child)

	childID := child.GetID()
	grandchild := &deck.Deck{}
	grandchild.SetName("Grandchild")
	grandchild.SetUserID(userID)
	grandchild.SetParentID(&childID)
	grandchild.SetOptionsJSON("{}")
	_ = deckRepo.Save(ctx, userID, grandchild)

	// Delete Parent
	err := deckRepo.Delete(ctx, userID, parentID)
	require.NoError(t, err)

	// Verify all are soft-deleted
	found, err := deckRepo.FindByID(ctx, userID, parentID)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)

	found, err = deckRepo.FindByID(ctx, userID, childID)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)

	found, err = deckRepo.FindByID(ctx, userID, grandchild.GetID())
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)
}

func TestDeckRepository_Exists(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_exists")

	deckEntity := &deck.Deck{}
	deckEntity.SetName("Existing Deck")
	deckEntity.SetUserID(userID)
	deckEntity.SetOptionsJSON("{}")
	_ = deckRepo.Save(ctx, userID, deckEntity)

	exists, err := deckRepo.Exists(ctx, userID, "Existing Deck", nil)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = deckRepo.Exists(ctx, userID, "Non Exists", nil)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestDeckRepository_UniqueConstraints(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_unique")

	t.Run("Root Name Conflict", func(t *testing.T) {
		name := "Root Deck"
		d1 := &deck.Deck{}
		d1.SetName(name)
		d1.SetUserID(userID)
		d1.SetOptionsJSON("{}")
		err := deckRepo.Save(ctx, userID, d1)
		require.NoError(t, err)

		d2 := &deck.Deck{}
		d2.SetName(name)
		d2.SetUserID(userID)
		d2.SetOptionsJSON("{}")
		err = deckRepo.Save(ctx, userID, d2)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists at this level")
	})

	t.Run("Child Name Conflict", func(t *testing.T) {
		parent := &deck.Deck{}
		parent.SetName("Parent")
		parent.SetUserID(userID)
		parent.SetOptionsJSON("{}")
		err := deckRepo.Save(ctx, userID, parent)
		require.NoError(t, err)
		parentID := parent.GetID()

		name := "Child Deck"
		c1 := &deck.Deck{}
		c1.SetName(name)
		c1.SetUserID(userID)
		c1.SetParentID(&parentID)
		c1.SetOptionsJSON("{}")
		err = deckRepo.Save(ctx, userID, c1)
		require.NoError(t, err)

		c2 := &deck.Deck{}
		c2.SetName(name)
		c2.SetUserID(userID)
		c2.SetParentID(&parentID)
		c2.SetOptionsJSON("{}")
		err = deckRepo.Save(ctx, userID, c2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists at this level")
	})

	t.Run("Same Name Different Parents Allowed", func(t *testing.T) {
		p1 := &deck.Deck{}
		p1.SetName("P1")
		p1.SetUserID(userID)
		p1.SetOptionsJSON("{}")
		deckRepo.Save(ctx, userID, p1)
		id1 := p1.GetID()

		p2 := &deck.Deck{}
		p2.SetName("P2")
		p2.SetUserID(userID)
		p2.SetOptionsJSON("{}")
		deckRepo.Save(ctx, userID, p2)
		id2 := p2.GetID()

		name := "Common Name"
		c1 := &deck.Deck{}
		c1.SetName(name)
		c1.SetUserID(userID)
		c1.SetParentID(&id1)
		c1.SetOptionsJSON("{}")
		err := deckRepo.Save(ctx, userID, c1)
		assert.NoError(t, err)

		c2 := &deck.Deck{}
		c2.SetName(name)
		c2.SetUserID(userID)
		c2.SetParentID(&id2)
		c2.SetOptionsJSON("{}")
		err = deckRepo.Save(ctx, userID, c2)
		assert.NoError(t, err)
	})
}

func TestDeckRepository_FindByUserID_Search(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deck_search")

	// Create multiple decks with different names
	decks := []struct {
		name string
	}{
		{"Math Deck"},
		{"Mathematics Advanced"},
		{"Science Deck"},
		{"Math Basics"},
		{"History"},
	}

	for _, d := range decks {
		deckEntity := &deck.Deck{}
		deckEntity.SetName(d.name)
		deckEntity.SetUserID(userID)
		deckEntity.SetOptionsJSON("{}")
		deckEntity.SetCreatedAt(time.Now())
		deckEntity.SetUpdatedAt(time.Now())
		_ = deckRepo.Save(ctx, userID, deckEntity)
	}

	t.Run("Search with Match", func(t *testing.T) {
		results, err := deckRepo.FindByUserID(ctx, userID, "Math")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3) // Should find "Math Deck", "Mathematics Advanced", "Math Basics"
		
		// Verify all results contain "Math" (case-insensitive)
		for _, d := range results {
			assert.Contains(t, d.GetName(), "Math")
		}
	})

	t.Run("Search Case-Insensitive", func(t *testing.T) {
		results, err := deckRepo.FindByUserID(ctx, userID, "math")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3)
		
		// Verify case-insensitive matching
		for _, d := range results {
			assert.Contains(t, d.GetName(), "Math")
		}
	})

	t.Run("Search with Partial Match", func(t *testing.T) {
		results, err := deckRepo.FindByUserID(ctx, userID, "Science")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Science Deck", results[0].GetName())
	})

	t.Run("Search with No Matches", func(t *testing.T) {
		results, err := deckRepo.FindByUserID(ctx, userID, "NonExistentDeck12345")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Search with Empty String", func(t *testing.T) {
		results, err := deckRepo.FindByUserID(ctx, userID, "")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 5) // Should return all decks
	})

	t.Run("Cross-User Isolation", func(t *testing.T) {
		// Create another user with a deck that matches search
		userID2, _ := createTestUser(t, ctx, userRepo, "deck_search_user2")
		deckEntity := &deck.Deck{}
		deckEntity.SetName("Math Deck")
		deckEntity.SetUserID(userID2)
		deckEntity.SetOptionsJSON("{}")
		deckEntity.SetCreatedAt(time.Now())
		deckEntity.SetUpdatedAt(time.Now())
		_ = deckRepo.Save(ctx, userID2, deckEntity)

		// User 1 should not see User 2's deck
		results, err := deckRepo.FindByUserID(ctx, userID, "Math")
		require.NoError(t, err)
		for _, d := range results {
			assert.Equal(t, userID, d.GetUserID(), "User 1 should not see User 2's decks")
		}
	})
}

