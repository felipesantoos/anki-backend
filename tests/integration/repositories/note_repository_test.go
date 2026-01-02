package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deckEntity "github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	searchdomain "github.com/felipesantos/anki-backend/core/domain/services/search"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	noteService "github.com/felipesantos/anki-backend/core/services/note"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/database"
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

func TestNoteRepository_FindByAdvancedSearch_Regex(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	parser := searchdomain.NewParser()

	userID, _ := createTestUser(t, ctx, userRepo, "note_regex_search")

	// Create note type with Front and Back fields
	noteType, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Template 1"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteTypeRepo.Save(ctx, userID, noteType)
	require.NoError(t, err)
	noteTypeID := noteType.GetID()

	// Create test notes with different field values
	counter := 0
	createNote := func(front, back string) int64 {
		counter++
		guid, err := valueobjects.NewGUID(fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", counter))
		require.NoError(t, err)

		noteEntity, err := note.NewBuilder().
			WithID(0).
			WithUserID(userID).
			WithGUID(guid).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(fmt.Sprintf(`{"Front":"%s","Back":"%s"}`, front, back)).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		require.NoError(t, err)

		err = noteRepo.Save(ctx, userID, noteEntity)
		require.NoError(t, err)
		return noteEntity.GetID()
	}

	// Create notes: a1, b1, c1 in Front; 123, 456 in Back; hello world in Front
	_ = createNote("a1", "test")
	_ = createNote("b1", "test")
	_ = createNote("c1", "test")
	_ = createNote("test", "123")
	_ = createNote("test", "456")
	_ = createNote("hello world", "test")

	t.Run("Basic_Regex", func(t *testing.T) {
		query, err := parser.Parse("re:hello.*world")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.Len(t, notes, 1)
		assert.Contains(t, notes[0].GetFieldsJSON(), "hello world")
	})

	t.Run("Field_Regex_Front", func(t *testing.T) {
		query, err := parser.Parse("front:re:[a-c]1")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.Len(t, notes, 3) // Should match a1, b1, c1
		for _, n := range notes {
			fields := n.GetFieldsJSON()
			// Check if fields contain the expected values (JSON order and spacing may vary)
			assert.True(t, strings.Contains(fields, `"Front":"a1"`) || strings.Contains(fields, `"Front": "a1"`) ||
				strings.Contains(fields, `"Front":"b1"`) || strings.Contains(fields, `"Front": "b1"`) ||
				strings.Contains(fields, `"Front":"c1"`) || strings.Contains(fields, `"Front": "c1"`), "Should contain a1, b1, or c1 in Front field")
			assert.True(t, strings.Contains(fields, `"Back":"test"`) || strings.Contains(fields, `"Back": "test"`), "Should contain Back:test")
		}
	})

	t.Run("Field_Regex_Back", func(t *testing.T) {
		query, err := parser.Parse("back:re:\\d{3}")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.Len(t, notes, 2) // Should match 123 and 456
		for _, n := range notes {
			fields := n.GetFieldsJSON()
			// Check if fields contain the expected values (JSON order and spacing may vary)
			assert.True(t, strings.Contains(fields, `"Front":"test"`) || strings.Contains(fields, `"Front": "test"`), "Should contain Front:test")
			assert.True(t, strings.Contains(fields, `"Back":"123"`) || strings.Contains(fields, `"Back": "123"`) ||
				strings.Contains(fields, `"Back":"456"`) || strings.Contains(fields, `"Back": "456"`), "Should contain Back:123 or Back:456")
		}
	})

	t.Run("Invalid_Regex", func(t *testing.T) {
		query, err := parser.Parse("re:[invalid")
		require.NoError(t, err) // Parser should accept it

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
		assert.Nil(t, notes)
	})
}

func TestNoteRepository_FindByAdvancedSearch_NoCombining(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	parser := searchdomain.NewParser()

	userID, _ := createTestUser(t, ctx, userRepo, "note_nocombining_search")

	// Create note type with Front and Back fields
	noteType, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Template 1"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteTypeRepo.Save(ctx, userID, noteType)
	require.NoError(t, err)
	noteTypeID := noteType.GetID()

	// Create test notes with accented text
	counter := 0
	createNote := func(front, back string) int64 {
		counter++
		guid, err := valueobjects.NewGUID(fmt.Sprintf("550e8400-e29b-41d4-a716-44665544%04d", counter))
		require.NoError(t, err)

		noteEntity, err := note.NewBuilder().
			WithID(0).
			WithUserID(userID).
			WithGUID(guid).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(fmt.Sprintf(`{"Front":"%s","Back":"%s"}`, front, back)).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		require.NoError(t, err)

		err = noteRepo.Save(ctx, userID, noteEntity)
		require.NoError(t, err)
		return noteEntity.GetID()
	}

	// Create notes with accented characters
	_ = createNote("café", "coffee")
	_ = createNote("ação", "action")
	_ = createNote("über", "over")
	_ = createNote("naïve", "naive")
	_ = createNote("résumé", "resume")
	_ = createNote("São Paulo", "city")

	t.Run("Basic_NoCombining", func(t *testing.T) {
		query, err := parser.Parse("nc:cafe")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notes), 1, "Should find at least one note with 'café'")
		found := false
		for _, n := range notes {
			if strings.Contains(n.GetFieldsJSON(), "café") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find note with 'café'")
	})

	t.Run("NoCombining_Field_Front", func(t *testing.T) {
		query, err := parser.Parse("front:nc:acao")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notes), 1, "Should find at least one note with 'ação' in Front")
		found := false
		for _, n := range notes {
			if strings.Contains(n.GetFieldsJSON(), "ação") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find note with 'ação' in Front field")
	})

	t.Run("NoCombining_Field_Back", func(t *testing.T) {
		query, err := parser.Parse("back:nc:coffee")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notes), 1, "Should find at least one note with 'coffee' in Back")
	})

	t.Run("NoCombining_With_Wildcard", func(t *testing.T) {
		query, err := parser.Parse("nc:uber*")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notes), 1, "Should find at least one note with 'über'")
		found := false
		for _, n := range notes {
			if strings.Contains(n.GetFieldsJSON(), "über") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find note with 'über'")
	})

	t.Run("NoCombining_Exact_Phrase", func(t *testing.T) {
		query, err := parser.Parse(`nc:"Sao Paulo"`)
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notes), 1, "Should find at least one note with 'São Paulo'")
		found := false
		for _, n := range notes {
			if strings.Contains(n.GetFieldsJSON(), "São Paulo") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find note with 'São Paulo'")
	})

	t.Run("NoCombining_With_Tag", func(t *testing.T) {
		// Create a note with tag
		guid, err := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655449999")
		require.NoError(t, err)

		noteEntity, err := note.NewBuilder().
			WithID(0).
			WithUserID(userID).
			WithGUID(guid).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"résumé","Back":"document"}`).
			WithTags([]string{"vocabulary"}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		require.NoError(t, err)

		err = noteRepo.Save(ctx, userID, noteEntity)
		require.NoError(t, err)

		query, err := parser.Parse("nc:resume tag:vocabulary")
		require.NoError(t, err)

		notes, err := noteRepo.FindByAdvancedSearch(ctx, userID, query, 100, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notes), 1, "Should find note with 'résumé' and tag 'vocabulary'")
		found := false
		for _, n := range notes {
			if strings.Contains(n.GetFieldsJSON(), "résumé") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find note with 'résumé'")
	})
}

func TestNoteService_Copy(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "note_copy_service")

	// Create note type
	noteType, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"},{"name":"Card 2"}]`). // 2 card types
		WithTemplatesJSON(`[{"name":"Template 1"},{"name":"Template 2"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteTypeRepo.Save(ctx, userID, noteType)
	require.NoError(t, err)
	noteTypeID := noteType.GetID()

	// Create deck
	defaultDeckID, err := deckRepo.CreateDefaultDeck(ctx, userID)
	require.NoError(t, err)
	deckID := defaultDeckID

	// Create original note with service
	tm := database.NewTransactionManager(db.DB)
	service := noteService.NewNoteService(noteRepo, cardRepo, noteTypeRepo, deckRepo, tm)
	originalNote, err := service.Create(ctx, userID, noteTypeID, deckID, `{"Front":"Original Question","Back":"Original Answer"}`, []string{"tag1", "tag2"})
	require.NoError(t, err)
	originalNoteID := originalNote.GetID()

	// Mark original note
	originalNote.Mark()
	err = noteRepo.Update(ctx, userID, originalNoteID, originalNote)
	require.NoError(t, err)

	t.Run("Success with all options", func(t *testing.T) {
		copiedNote, err := service.Copy(ctx, userID, originalNoteID, nil, true, true)
		require.NoError(t, err)
		assert.NotNil(t, copiedNote)
		assert.NotEqual(t, originalNoteID, copiedNote.GetID(), "Copy should have different ID")
		assert.NotEqual(t, originalNote.GetGUID().Value(), copiedNote.GetGUID().Value(), "Copy should have different GUID")
		assert.Equal(t, originalNote.GetNoteTypeID(), copiedNote.GetNoteTypeID())
		// Compare JSON fields (order may vary, so parse and compare)
		var originalFields, copiedFields map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(originalNote.GetFieldsJSON()), &originalFields))
		require.NoError(t, json.Unmarshal([]byte(copiedNote.GetFieldsJSON()), &copiedFields))
		assert.Equal(t, originalFields, copiedFields, "Fields should be copied")
		assert.Equal(t, originalNote.GetTags(), copiedNote.GetTags(), "Tags should be copied")
		assert.False(t, copiedNote.GetMarked(), "Copy should not inherit marked status")

		// Verify cards were created
		cards, err := cardRepo.FindByNoteID(ctx, userID, copiedNote.GetID())
		require.NoError(t, err)
		assert.Equal(t, 2, len(cards), "Should create 2 cards (one for each card type)")
		for _, c := range cards {
			assert.Equal(t, deckID, c.GetDeckID(), "Cards should be in same deck as original")
		}
	})

	t.Run("Success without tags", func(t *testing.T) {
		copiedNote, err := service.Copy(ctx, userID, originalNoteID, nil, false, false)
		require.NoError(t, err)
		assert.NotNil(t, copiedNote)
		assert.Empty(t, copiedNote.GetTags(), "Tags should not be copied")
		// Compare JSON fields (order may vary, so parse and compare)
		var originalFields, copiedFields map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(originalNote.GetFieldsJSON()), &originalFields))
		require.NoError(t, json.Unmarshal([]byte(copiedNote.GetFieldsJSON()), &copiedFields))
		assert.Equal(t, originalFields, copiedFields, "Fields should still be copied")
	})

	t.Run("Success with different deck", func(t *testing.T) {
		// Create another deck with a different name
		deck2Entity, err := deckEntity.NewBuilder().
			WithUserID(userID).
			WithName("Test Deck 2").
			WithOptionsJSON("{}").
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		require.NoError(t, err)
		err = deckRepo.Save(ctx, userID, deck2Entity)
		require.NoError(t, err)
		deck2ID := deck2Entity.GetID()

		copiedNote, err := service.Copy(ctx, userID, originalNoteID, &deck2ID, true, true)
		require.NoError(t, err)
		assert.NotNil(t, copiedNote)

		// Verify cards are in the new deck
		cards, err := cardRepo.FindByNoteID(ctx, userID, copiedNote.GetID())
		require.NoError(t, err)
		for _, c := range cards {
			assert.Equal(t, deck2ID, c.GetDeckID(), "Cards should be in the specified deck")
		}
	})

	t.Run("Success same deck", func(t *testing.T) {
		copiedNote, err := service.Copy(ctx, userID, originalNoteID, nil, true, true)
		require.NoError(t, err)
		assert.NotNil(t, copiedNote)

		// Verify cards are in the same deck as original
		cards, err := cardRepo.FindByNoteID(ctx, userID, copiedNote.GetID())
		require.NoError(t, err)
		for _, c := range cards {
			assert.Equal(t, deckID, c.GetDeckID(), "Cards should be in same deck as original when deckID is nil")
		}
	})

	t.Run("Generates new cards", func(t *testing.T) {
		copiedNote, err := service.Copy(ctx, userID, originalNoteID, nil, true, true)
		require.NoError(t, err)

		// Get original cards
		originalCards, err := cardRepo.FindByNoteID(ctx, userID, originalNoteID)
		require.NoError(t, err)

		// Get copied cards
		copiedCards, err := cardRepo.FindByNoteID(ctx, userID, copiedNote.GetID())
		require.NoError(t, err)

		assert.Equal(t, len(originalCards), len(copiedCards), "Should have same number of cards")
		for _, copiedCard := range copiedCards {
			assert.Equal(t, copiedNote.GetID(), copiedCard.GetNoteID(), "Card should reference copied note")
			assert.NotEqual(t, originalNoteID, copiedCard.GetNoteID(), "Card should not reference original note")
		}
	})

	t.Run("New GUID", func(t *testing.T) {
		copiedNote, err := service.Copy(ctx, userID, originalNoteID, nil, true, true)
		require.NoError(t, err)
		assert.NotEqual(t, originalNote.GetGUID().Value(), copiedNote.GetGUID().Value(), "Copy should have new GUID")
	})

	t.Run("Cross-user isolation", func(t *testing.T) {
		// Create another user
		userID2, _ := createTestUser(t, ctx, userRepo, "note_copy_user2")

		// Try to copy User A's note as User B
		_, err := service.Copy(ctx, userID2, originalNoteID, nil, true, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found", "Should not find note from other user")
	})

	t.Run("Original note unchanged", func(t *testing.T) {
		// Get original note before copy
		originalBefore, err := noteRepo.FindByID(ctx, userID, originalNoteID)
		require.NoError(t, err)

		// Copy note
		_, err = service.Copy(ctx, userID, originalNoteID, nil, true, true)
		require.NoError(t, err)

		// Get original note after copy
		originalAfter, err := noteRepo.FindByID(ctx, userID, originalNoteID)
		require.NoError(t, err)

		// Verify original note is unchanged
		assert.Equal(t, originalBefore.GetID(), originalAfter.GetID())
		assert.Equal(t, originalBefore.GetGUID().Value(), originalAfter.GetGUID().Value())
		assert.Equal(t, originalBefore.GetFieldsJSON(), originalAfter.GetFieldsJSON())
		assert.Equal(t, originalBefore.GetTags(), originalAfter.GetTags())
	})

	t.Run("Note not found", func(t *testing.T) {
		_, err := service.Copy(ctx, userID, 999999, nil, true, true)
		assert.Error(t, err)
		// Service converts ownership.ErrResourceNotFound to "note not found"
		assert.Contains(t, err.Error(), "note not found")
	})

	t.Run("Deck not found", func(t *testing.T) {
		invalidDeckID := int64(999999)
		_, err := service.Copy(ctx, userID, originalNoteID, &invalidDeckID, true, true)
		assert.Error(t, err)
		// Service converts ownership.ErrResourceNotFound to "deck not found"
		assert.Contains(t, err.Error(), "deck not found")
	})
}

func TestNoteRepository_FindDuplicatesByField(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteRepo := repositories.NewNoteRepository(db.DB)
	cardRepo := repositories.NewCardRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)
	deckRepo := repositories.NewDeckRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "find_duplicates_user")

	// Create note type
	noteType, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Template 1"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = noteTypeRepo.Save(ctx, userID, noteType)
	require.NoError(t, err)
	noteTypeID := noteType.GetID()

	// Create deck
	defaultDeckID, err := deckRepo.CreateDefaultDeck(ctx, userID)
	require.NoError(t, err)
	deckID := defaultDeckID

	// Create service for creating notes
	tm := database.NewTransactionManager(db.DB)
	service := noteService.NewNoteService(noteRepo, cardRepo, noteTypeRepo, deckRepo, tm)

	// Create duplicate notes (3 notes with "Hello" in Front field)
	note1, err := service.Create(ctx, userID, noteTypeID, deckID, `{"Front":"Hello","Back":"World1"}`, []string{})
	require.NoError(t, err)

	note2, err := service.Create(ctx, userID, noteTypeID, deckID, `{"Front":"Hello","Back":"Different"}`, []string{})
	require.NoError(t, err)

	note3, err := service.Create(ctx, userID, noteTypeID, deckID, `{"Front":"Goodbye","Back":"World2"}`, []string{})
	require.NoError(t, err)

	note4, err := service.Create(ctx, userID, noteTypeID, deckID, `{"Front":"Hello","Back":"Another"}`, []string{})
	require.NoError(t, err)

	t.Run("Success find duplicates by field", func(t *testing.T) {
		groups, err := noteRepo.FindDuplicatesByField(ctx, userID, &noteTypeID, "Front")
		require.NoError(t, err)
		assert.Len(t, groups, 1, "Should find one duplicate group")
		assert.Equal(t, "Hello", groups[0].FieldValue)
		assert.Len(t, groups[0].Notes, 3, "Should have 3 notes with 'Hello' in Front field")
		
		// Verify note IDs
		noteIDs := make(map[int64]bool)
		for _, n := range groups[0].Notes {
			noteIDs[n.ID] = true
		}
		assert.True(t, noteIDs[note1.GetID()])
		assert.True(t, noteIDs[note2.GetID()])
		assert.True(t, noteIDs[note4.GetID()])
		assert.False(t, noteIDs[note3.GetID()])
	})

	t.Run("Success no duplicates found", func(t *testing.T) {
		// All Back values are unique (World1, Different, World2, Another), so no duplicates should be found
		groups, err := noteRepo.FindDuplicatesByField(ctx, userID, &noteTypeID, "Back")
		require.NoError(t, err)
		assert.Empty(t, groups, "Should find no duplicates in Back field")
	})

	t.Run("Success without note type filter", func(t *testing.T) {
		groups, err := noteRepo.FindDuplicatesByField(ctx, userID, nil, "Front")
		require.NoError(t, err)
		assert.Len(t, groups, 1, "Should find one duplicate group even without note type filter")
	})

	t.Run("Cross-user isolation", func(t *testing.T) {
		userID2, _ := createTestUser(t, ctx, userRepo, "find_duplicates_user2")
		
		// Create note for user 2 with same field value
		noteType2, err := notetype.NewBuilder().
			WithID(0).
			WithUserID(userID2).
			WithName("Basic").
			WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
			WithCardTypesJSON(`[{"name":"Card 1"}]`).
			WithTemplatesJSON(`[{"name":"Template 1"}]`).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()
		require.NoError(t, err)
		err = noteTypeRepo.Save(ctx, userID2, noteType2)
		require.NoError(t, err)

		defaultDeckID2, err := deckRepo.CreateDefaultDeck(ctx, userID2)
		require.NoError(t, err)

		_, err = service.Create(ctx, userID2, noteType2.GetID(), defaultDeckID2, `{"Front":"Hello","Back":"World"}`, []string{})
		require.NoError(t, err)

		// User 1 should not see User 2's duplicates
		groups, err := noteRepo.FindDuplicatesByField(ctx, userID, nil, "Front")
		require.NoError(t, err)
		// Should still only find 3 notes (from user 1), not 4
		if len(groups) > 0 {
			assert.LessOrEqual(t, len(groups[0].Notes), 3, "Should not include User 2's notes")
		}
	})

	t.Run("Empty field name", func(t *testing.T) {
		groups, err := noteRepo.FindDuplicatesByField(ctx, userID, nil, "")
		require.NoError(t, err)
		assert.Empty(t, groups)
	})
}

