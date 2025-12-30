package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestNoteRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	// Create test user
	userID, _ := createTestUser(t, ctx, userRepo, "note_save")

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
	noteTypeID := noteType.GetID()

	// Create note
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteTypeID).
		WithFieldsJSON(`{"Front":"Hello"}`).
		WithTags([]string{"tag1", "tag2"}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)
	assert.Greater(t, noteEntity.GetID(), int64(0))
}

func TestNoteRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "note_find")

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

	// Create and save note
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"Test"}`).
		WithTags([]string{"test"}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)
	noteID := noteEntity.GetID()

	// Find by ID
	found, err := noteRepo.FindByID(ctx, userID, noteID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, noteID, found.GetID())
	assert.Equal(t, userID, found.GetUserID())
}

func TestNoteRepository_FindByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "note_notfound")

	found, err := noteRepo.FindByID(ctx, userID, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)
}

func TestNoteRepository_FindByGUID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "note_guid")

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

	// Create note with specific GUID
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440002")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"GUID Test"}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)

	// Find by GUID
	found, err := noteRepo.FindByGUID(ctx, userID, guid.Value())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, guid.Value(), found.GetGUID().Value())
}

func TestNoteRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "note_update")

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
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440003")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"Original"}`).
		WithTags([]string{"original"}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)
	noteID := noteEntity.GetID()

	// Update note
	noteEntity.SetFieldsJSON(`{"Front":"Updated"}`)
	noteEntity.SetTags([]string{"updated"})
	noteEntity.SetMarked(true)
	err = noteRepo.Update(ctx, userID, noteID, noteEntity)
	require.NoError(t, err)

	// Verify update
	updated, err := noteRepo.FindByID(ctx, userID, noteID)
	require.NoError(t, err)
	assert.JSONEq(t, `{"Front":"Updated"}`, updated.GetFieldsJSON())
	assert.Equal(t, []string{"updated"}, updated.GetTags())
	assert.True(t, updated.GetMarked())
}

func TestNoteRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "note_delete")

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
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440004")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"Delete Test"}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)
	noteID := noteEntity.GetID()

	// Delete (soft delete)
	err = noteRepo.Delete(ctx, userID, noteID)
	require.NoError(t, err)

	// Verify soft delete
	found, err := noteRepo.FindByID(ctx, userID, noteID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found) // Should not find soft-deleted note
}

func TestNoteRepository_Exists(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "note_exists")

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
	guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440005")
	require.NoError(t, err)

	noteEntity, err := note.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithGUID(guid).
		WithNoteTypeID(noteType.GetID()).
		WithFieldsJSON(`{"Front":"Exists Test"}`).
		WithTags([]string{}).
		WithMarked(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteRepo.Save(ctx, userID, noteEntity)
	require.NoError(t, err)
	noteID := noteEntity.GetID()

	// Test exists
	exists, err := noteRepo.Exists(ctx, userID, noteID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test not exists
	exists, err = noteRepo.Exists(ctx, userID, 99999)
	require.NoError(t, err)
	assert.False(t, exists)
}

