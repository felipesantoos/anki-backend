package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestNoteTypeRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_save")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"qfmt":"{{Front}}","afmt":"{{Back}}"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)
	assert.Greater(t, noteTypeEntity.GetID(), int64(0))
}

func TestNoteTypeRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_find")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Cloze").
		WithFieldsJSON(`[{"name":"Text"}]`).
		WithCardTypesJSON(`[{"name":"Cloze"}]`).
		WithTemplatesJSON(`[{"qfmt":"{{Text}}","afmt":"{{Text}}"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)
	noteTypeID := noteTypeEntity.GetID()

	found, err := noteTypeRepo.FindByID(ctx, userID, noteTypeID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, noteTypeID, found.GetID())
	assert.Equal(t, "Cloze", found.GetName())
}

func TestNoteTypeRepository_FindByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_name")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Unique Note Type").
		WithFieldsJSON(`[{"name":"Field"}]`).
		WithCardTypesJSON(`[{"name":"Card"}]`).
		WithTemplatesJSON(`[{"qfmt":"{{Field}}","afmt":"{{Field}}"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)

	found, err := noteTypeRepo.FindByName(ctx, userID, "Unique Note Type")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Unique Note Type", found.GetName())
}

func TestNoteTypeRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_delete")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("To Delete").
		WithFieldsJSON(`[{"name":"Field"}]`).
		WithCardTypesJSON(`[{"name":"Card"}]`).
		WithTemplatesJSON(`[{"qfmt":"{{Field}}","afmt":"{{Field}}"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)
	noteTypeID := noteTypeEntity.GetID()

	err = noteTypeRepo.Delete(ctx, userID, noteTypeID)
	require.NoError(t, err)

	found, err := noteTypeRepo.FindByID(ctx, userID, noteTypeID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)
}

func TestNoteTypeRepository_FindByUserID_Search(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_search")
	otherUserID, _ := createTestUser(t, ctx, userRepo, "notetype_search_other_user")

	// Helper function to create note type entity
	createNoteTypeEntity := func(t *testing.T, userID int64, name string) *notetype.NoteType {
		nt, err := notetype.NewBuilder().
			WithID(0).
			WithUserID(userID).
			WithName(name).
			WithFieldsJSON(`[{"name":"Field"}]`).
			WithCardTypesJSON(`[{"name":"Card"}]`).
			WithTemplatesJSON(`[{"qfmt":"{{Field}}","afmt":"{{Field}}"}]`).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		require.NoError(t, err)
		return nt
	}

	// Create note types for userID
	nt1 := createNoteTypeEntity(t, userID, "My Basic Note Type")
	nt2 := createNoteTypeEntity(t, userID, "My Cloze Note Type")
	nt3 := createNoteTypeEntity(t, userID, "Another Note Type")
	_ = noteTypeRepo.Save(ctx, userID, nt1)
	_ = noteTypeRepo.Save(ctx, userID, nt2)
	_ = noteTypeRepo.Save(ctx, userID, nt3)

	// Create a note type for otherUserID with a similar name
	ntOther := createNoteTypeEntity(t, otherUserID, "My Basic Note Type")
	_ = noteTypeRepo.Save(ctx, otherUserID, ntOther)

	t.Run("Search with Match", func(t *testing.T) {
		noteTypes, err := noteTypeRepo.FindByUserID(ctx, userID, "Basic")
		require.NoError(t, err)
		assert.Len(t, noteTypes, 1)
		assert.Equal(t, "My Basic Note Type", noteTypes[0].GetName())
	})

	t.Run("Case-Insensitive Search", func(t *testing.T) {
		noteTypes, err := noteTypeRepo.FindByUserID(ctx, userID, "basic")
		require.NoError(t, err)
		assert.Len(t, noteTypes, 1)
		assert.Equal(t, "My Basic Note Type", noteTypes[0].GetName())
	})

	t.Run("Partial Matching", func(t *testing.T) {
		noteTypes, err := noteTypeRepo.FindByUserID(ctx, userID, "Note Type")
		require.NoError(t, err)
		assert.Len(t, noteTypes, 3) // "My Basic Note Type", "My Cloze Note Type", "Another Note Type"
	})

	t.Run("No Results", func(t *testing.T) {
		noteTypes, err := noteTypeRepo.FindByUserID(ctx, userID, "NonExistent")
		require.NoError(t, err)
		assert.Empty(t, noteTypes)
	})

	t.Run("Empty Search String", func(t *testing.T) {
		noteTypes, err := noteTypeRepo.FindByUserID(ctx, userID, "")
		require.NoError(t, err)
		assert.Len(t, noteTypes, 3) // All note types for userID
	})

	t.Run("Cross-User Isolation", func(t *testing.T) {
		noteTypes, err := noteTypeRepo.FindByUserID(ctx, otherUserID, "Basic")
		require.NoError(t, err)
		assert.Len(t, noteTypes, 1)
		assert.Equal(t, "My Basic Note Type", noteTypes[0].GetName())
		assert.Equal(t, otherUserID, noteTypes[0].GetUserID()) // Ensure it's the other user's note type
	})
}

