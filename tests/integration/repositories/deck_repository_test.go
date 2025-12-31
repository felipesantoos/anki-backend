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
	allDecks, _ := deckRepo.FindByUserID(ctx, userID)
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

