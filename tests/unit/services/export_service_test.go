package services

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	exportSvc "github.com/felipesantos/anki-backend/core/services/export"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestExportService_ExportNotes_TextFormat(t *testing.T) {
	mockDeckRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteRepo := new(MockNoteRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockMediaRepo := new(MockMediaRepository)
	service := exportSvc.NewExportService(mockDeckRepo, mockCardRepo, mockNoteRepo, mockNoteTypeRepo, mockMediaRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success - Text format without scheduling", func(t *testing.T) {
		noteIDs := []int64{1, 2}
		noteTypeID := int64(10)

		// Create test notes
		guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
		note1, _ := note.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithGUID(guid1).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Hello","Back":"World"}`).
			WithTags([]string{"tag1"}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		guid2, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440002")
		note2, _ := note.NewBuilder().
			WithID(2).
			WithUserID(userID).
			WithGUID(guid2).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Test","Back":"Card"}`).
			WithTags([]string{"tag2"}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		notes := []*note.Note{note1, note2}

		// Create note type
		noteType, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithName("Basic").
			WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
			WithCardTypesJSON(`[{"name":"Card 1"}]`).
			WithTemplatesJSON(`[{"name":"Template 1"}]`).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		mockNoteRepo.On("FindByIDs", ctx, userID, noteIDs).Return(notes, nil).Once()
		mockNoteTypeRepo.On("FindByID", ctx, userID, noteTypeID).Return(noteType, nil).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, noteIDs, "text", false, false)
		require.NoError(t, err)
		assert.Greater(t, size, int64(0))
		assert.Equal(t, "notes_export.txt", filename)

		// Read and verify content
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Contains(t, string(data), "GUID")
		assert.Contains(t, string(data), "Front")
		assert.Contains(t, string(data), "Back")
		assert.Contains(t, string(data), "Tags")
	})

	t.Run("Success - Text format with scheduling", func(t *testing.T) {
		noteIDs := []int64{1}
		noteTypeID := int64(10)
		deckID := int64(20)

		guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
		note1, _ := note.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithGUID(guid1).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Hello"}`).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		notes := []*note.Note{note1}

		card1, _ := card.NewBuilder().
			WithID(100).
			WithNoteID(1).
			WithCardTypeID(0).
			WithDeckID(deckID).
			WithDue(time.Now().Unix() * 1000).
			WithInterval(1).
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

		cards := []*card.Card{card1}

		noteType, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithName("Basic").
			WithFieldsJSON(`[{"name":"Front"}]`).
			WithCardTypesJSON(`[{"name":"Card 1"}]`).
			WithTemplatesJSON(`[{"name":"Template 1"}]`).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		deck1, _ := deck.NewBuilder().
			WithID(deckID).
			WithUserID(userID).
			WithName("Test Deck").
			WithOptionsJSON("{}").
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		mockNoteRepo.On("FindByIDs", ctx, userID, noteIDs).Return(notes, nil).Once()
		mockCardRepo.On("FindByNoteIDs", ctx, userID, noteIDs).Return(cards, nil).Once()
		mockNoteTypeRepo.On("FindByID", ctx, userID, noteTypeID).Return(noteType, nil).Once()
		mockDeckRepo.On("FindByID", ctx, userID, deckID).Return(deck1, nil).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, noteIDs, "text", false, true)
		require.NoError(t, err)
		assert.Greater(t, size, int64(0))
		assert.Equal(t, "notes_export.txt", filename)

		// Read and verify content includes card info
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Contains(t, string(data), "CARD")
	})
}

func TestExportService_ExportNotes_APKGFormat(t *testing.T) {
	mockDeckRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteRepo := new(MockNoteRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockMediaRepo := new(MockMediaRepository)
	service := exportSvc.NewExportService(mockDeckRepo, mockCardRepo, mockNoteRepo, mockNoteTypeRepo, mockMediaRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("APKG format - Not yet implemented", func(t *testing.T) {
		noteIDs := []int64{1}
		noteTypeID := int64(10)

		guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
		note1, _ := note.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithGUID(guid1).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Hello"}`).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		notes := []*note.Note{note1}

		noteType, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithName("Basic").
			WithFieldsJSON(`[{"name":"Front"}]`).
			WithCardTypesJSON(`[{"name":"Card 1"}]`).
			WithTemplatesJSON(`[{"name":"Template 1"}]`).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		mockNoteRepo.On("FindByIDs", ctx, userID, noteIDs).Return(notes, nil).Once()
		mockNoteTypeRepo.On("FindByID", ctx, userID, noteTypeID).Return(noteType, nil).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, noteIDs, "apkg", false, false)
		// APKG generation requires SQLite driver which is not yet implemented
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SQLite")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})
}

func TestExportService_ExportNotes_Validation(t *testing.T) {
	mockDeckRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteRepo := new(MockNoteRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockMediaRepo := new(MockMediaRepository)
	service := exportSvc.NewExportService(mockDeckRepo, mockCardRepo, mockNoteRepo, mockNoteTypeRepo, mockMediaRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Empty note IDs", func(t *testing.T) {
		reader, size, filename, err := service.ExportNotes(ctx, userID, []int64{}, "text", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})

	t.Run("Invalid format", func(t *testing.T) {
		reader, size, filename, err := service.ExportNotes(ctx, userID, []int64{1}, "invalid", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})

	t.Run("No notes found", func(t *testing.T) {
		mockNoteRepo.On("FindByIDs", ctx, userID, []int64{1}).Return([]*note.Note{}, nil).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, []int64{1}, "text", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no notes found")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})

	t.Run("Note not found or access denied", func(t *testing.T) {
		noteIDs := []int64{1, 2}
		noteTypeID := int64(10)

		guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
		note1, _ := note.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithGUID(guid1).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Hello"}`).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		// Only note 1 is returned, note 2 is missing (access denied)
		notes := []*note.Note{note1}

		mockNoteRepo.On("FindByIDs", ctx, userID, noteIDs).Return(notes, nil).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, noteIDs, "text", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found or access denied")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})
}

func TestExportService_ExportNotes_MediaExtraction(t *testing.T) {
	mockDeckRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteRepo := new(MockNoteRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockMediaRepo := new(MockMediaRepository)
	service := exportSvc.NewExportService(mockDeckRepo, mockCardRepo, mockNoteRepo, mockNoteTypeRepo, mockMediaRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success - Text format with media", func(t *testing.T) {
		noteIDs := []int64{1}
		noteTypeID := int64(10)

		guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
		note1, _ := note.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithGUID(guid1).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"<img src=\"image.jpg\">"}`).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		notes := []*note.Note{note1}

		noteType, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithName("Basic").
			WithFieldsJSON(`[{"name":"Front"}]`).
			WithCardTypesJSON(`[{"name":"Card 1"}]`).
			WithTemplatesJSON(`[{"name":"Template 1"}]`).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		media1, _ := media.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithFilename("image.jpg").
			WithHash("abc123").
			WithSize(1024).
			WithMimeType("image/jpeg").
			WithStoragePath("/path/to/image.jpg").
			WithCreatedAt(time.Now()).
			Build()

		mockNoteRepo.On("FindByIDs", ctx, userID, noteIDs).Return(notes, nil).Once()
		mockNoteTypeRepo.On("FindByID", ctx, userID, noteTypeID).Return(noteType, nil).Once()
		mockMediaRepo.On("FindByFilename", ctx, userID, "image.jpg").Return(media1, nil).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, noteIDs, "text", true, false)
		require.NoError(t, err)
		assert.Greater(t, size, int64(0))
		assert.Equal(t, "notes_export.txt", filename)

		// Media extraction should not affect text export content
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Contains(t, string(data), "Front")
	})
}

func TestExportService_ExportNotes_ErrorHandling(t *testing.T) {
	mockDeckRepo := new(MockDeckRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteRepo := new(MockNoteRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockMediaRepo := new(MockMediaRepository)
	service := exportSvc.NewExportService(mockDeckRepo, mockCardRepo, mockNoteRepo, mockNoteTypeRepo, mockMediaRepo)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Repository error - FindByIDs", func(t *testing.T) {
		mockNoteRepo.On("FindByIDs", ctx, userID, []int64{1}).Return(nil, assert.AnError).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, []int64{1}, "text", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch notes")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})

	t.Run("Repository error - FindByNoteIDs (cards)", func(t *testing.T) {
		noteIDs := []int64{1}
		noteTypeID := int64(10)

		guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
		note1, _ := note.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithGUID(guid1).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Hello"}`).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		notes := []*note.Note{note1}

		noteType, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithName("Basic").
			WithFieldsJSON(`[{"name":"Front"}]`).
			WithCardTypesJSON(`[{"name":"Card 1"}]`).
			WithTemplatesJSON(`[{"name":"Template 1"}]`).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		mockNoteRepo.On("FindByIDs", ctx, userID, noteIDs).Return(notes, nil).Once()
		mockCardRepo.On("FindByNoteIDs", ctx, userID, noteIDs).Return(nil, assert.AnError).Once()
		mockNoteTypeRepo.On("FindByID", ctx, userID, noteTypeID).Return(noteType, nil).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, noteIDs, "text", false, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch cards")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})

	t.Run("Repository error - FindByID (note type)", func(t *testing.T) {
		noteIDs := []int64{1}
		noteTypeID := int64(10)

		guid1, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440001")
		note1, _ := note.NewBuilder().
			WithID(1).
			WithUserID(userID).
			WithGUID(guid1).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Hello"}`).
			WithTags([]string{}).
			WithMarked(false).
			WithCreatedAt(time.Now()).
			WithUpdatedAt(time.Now()).
			Build()

		notes := []*note.Note{note1}

		mockNoteRepo.On("FindByIDs", ctx, userID, noteIDs).Return(notes, nil).Once()
		mockNoteTypeRepo.On("FindByID", ctx, userID, noteTypeID).Return(nil, ownership.ErrResourceNotFound).Once()

		reader, size, filename, err := service.ExportNotes(ctx, userID, noteIDs, "text", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch note type")
		assert.Equal(t, int64(0), size)
		assert.Equal(t, "", filename)
		assert.Nil(t, reader)
	})
}

