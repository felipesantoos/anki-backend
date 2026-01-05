package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	noteSvc "github.com/felipesantos/anki-backend/core/services/note"
	"github.com/felipesantos/anki-backend/pkg/ownership"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoteService_Create(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success", func(t *testing.T) {
		noteTypeID := int64(10)
		deckID := int64(20)
		fields := "{\"Front\":\"Q\", \"Back\":\"A\"}"
		
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithCardTypesJSON("[{}, {}]"). // 2 card types
			Build()

		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Target Deck").Build()

		mockTM.ExpectTransaction()
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockDeckRepo.On("FindByID", mock.Anything, userID, deckID).Return(d, nil).Once()
		mockNoteRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*note.Note")).Return(nil).Run(func(args mock.Arguments) {
			n := args.Get(2).(*note.Note)
			n.SetID(100) // Set ID so card generation works
		}).Once()
		mockCardRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*card.Card")).Return(nil).Twice()

		result, err := service.Create(ctx, userID, noteTypeID, deckID, fields, []string{"tag1"})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Deck Not Found or Unauthorized", func(t *testing.T) {
		noteTypeID := int64(10)
		deckID := int64(404)
		
		nt, _ := notetype.NewBuilder().WithID(noteTypeID).WithUserID(userID).Build()

		mockTM.ExpectTransaction()
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockDeckRepo.On("FindByID", mock.Anything, userID, deckID).Return(nil, nil).Once()

		result, err := service.Create(ctx, userID, noteTypeID, deckID, "{}", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deck not found")
		assert.Nil(t, result)
	})

	t.Run("Generates GUID automatically", func(t *testing.T) {
		noteTypeID := int64(10)
		deckID := int64(20)
		fields := "{\"Front\":\"Q\", \"Back\":\"A\"}"
		
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithCardTypesJSON("[{}]"). // 1 card type
			Build()

		d, _ := deck.NewBuilder().WithID(deckID).WithUserID(userID).WithName("Target Deck").Build()

		mockTM.ExpectTransaction()
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockDeckRepo.On("FindByID", mock.Anything, userID, deckID).Return(d, nil).Once()
		mockNoteRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*note.Note")).Return(nil).Run(func(args mock.Arguments) {
			n := args.Get(2).(*note.Note)
			n.SetID(100) // Set ID so card generation works
			// Verify GUID is generated and valid
			guid := n.GetGUID()
			assert.False(t, guid.IsEmpty(), "GUID should not be empty")
			assert.NotEmpty(t, guid.Value(), "GUID should have a value")
			// Verify GUID format is valid UUID (RFC 4122)
			guidValue := guid.Value()
			assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, guidValue, "GUID should be in UUID format")
		}).Once()
		mockCardRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*card.Card")).Return(nil).Once()

		result, err := service.Create(ctx, userID, noteTypeID, deckID, fields, []string{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Verify GUID is present in result
		guid := result.GetGUID()
		assert.False(t, guid.IsEmpty(), "GUID should not be empty")
		assert.NotEmpty(t, guid.Value(), "GUID should have a value")
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestNoteService_FindAll(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Default - No Filters", func(t *testing.T) {
		filters := note.NoteFilters{}
		n1 := &note.Note{}; n1.SetID(1)
		n2 := &note.Note{}; n2.SetID(2)
		expectedNotes := []*note.Note{n1, n2}
		mockNoteRepo.On("FindByUserID", ctx, userID, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by DeckID", func(t *testing.T) {
		deckID := int64(10)
		filters := note.NoteFilters{DeckID: &deckID}
		n1 := &note.Note{}; n1.SetID(1)
		expectedNotes := []*note.Note{n1}
		mockNoteRepo.On("FindByDeckID", ctx, userID, deckID, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by NoteTypeID", func(t *testing.T) {
		noteTypeID := int64(20)
		filters := note.NoteFilters{NoteTypeID: &noteTypeID}
		n2 := &note.Note{}; n2.SetID(2)
		expectedNotes := []*note.Note{n2}
		mockNoteRepo.On("FindByNoteTypeID", ctx, userID, noteTypeID, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by Tags", func(t *testing.T) {
		tags := []string{"tag1"}
		filters := note.NoteFilters{Tags: tags}
		n3 := &note.Note{}; n3.SetID(3)
		expectedNotes := []*note.Note{n3}
		mockNoteRepo.On("FindByTags", ctx, userID, tags, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by Tags - Multiple Tags (OR logic)", func(t *testing.T) {
		tags := []string{"tag1", "tag2", "tag3"}
		filters := note.NoteFilters{Tags: tags}
		n1 := &note.Note{}; n1.SetID(1)
		n2 := &note.Note{}; n2.SetID(2)
		expectedNotes := []*note.Note{n1, n2}
		// Should find notes with ANY of the tags (OR logic)
		mockNoteRepo.On("FindByTags", ctx, userID, tags, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by Tags with Pagination", func(t *testing.T) {
		tags := []string{"tag1"}
		filters := note.NoteFilters{Tags: tags, Limit: 10, Offset: 20}
		n1 := &note.Note{}; n1.SetID(1)
		expectedNotes := []*note.Note{n1}
		mockNoteRepo.On("FindByTags", ctx, userID, tags, 10, 20).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by Tags - Empty Tags Array", func(t *testing.T) {
		tags := []string{}
		filters := note.NoteFilters{Tags: tags}
		// Empty tags should return empty results, not call FindByTags
		mockNoteRepo.On("FindByUserID", ctx, userID, 50, 0).Return([]*note.Note{}, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Tag Search Priority - Search Takes Precedence", func(t *testing.T) {
		searchText := "test"
		tags := []string{"tag1", "tag2"}
		filters := note.NoteFilters{Search: searchText, Tags: tags}
		n1 := &note.Note{}; n1.SetID(1)
		expectedNotes := []*note.Note{n1}
		// Search should be called, not FindByTags (Search has higher priority)
		mockNoteRepo.On("FindBySearch", ctx, userID, searchText, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by Search", func(t *testing.T) {
		searchText := "hello"
		filters := note.NoteFilters{Search: searchText}
		n4 := &note.Note{}; n4.SetID(4)
		expectedNotes := []*note.Note{n4}
		mockNoteRepo.On("FindBySearch", ctx, userID, searchText, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by Search with Pagination", func(t *testing.T) {
		searchText := "world"
		filters := note.NoteFilters{Search: searchText, Limit: 10, Offset: 20}
		mockNoteRepo.On("FindBySearch", ctx, userID, searchText, 10, 20).Return([]*note.Note{}, nil).Once()

		_, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Filter by Search - Empty Search Text", func(t *testing.T) {
		filters := note.NoteFilters{Search: ""}
		// Empty search should return empty results, not call FindBySearch
		mockNoteRepo.On("FindByUserID", ctx, userID, 50, 0).Return([]*note.Note{}, nil).Once()

		_, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Search Priority over Other Filters", func(t *testing.T) {
		searchText := "test"
		deckID := int64(10)
		filters := note.NoteFilters{Search: searchText, DeckID: &deckID}
		n5 := &note.Note{}; n5.SetID(5)
		expectedNotes := []*note.Note{n5}
		// Search should be called, not FindByDeckID
		mockNoteRepo.On("FindBySearch", ctx, userID, searchText, 50, 0).Return(expectedNotes, nil).Once()

		result, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Pagination", func(t *testing.T) {
		filters := note.NoteFilters{Limit: 10, Offset: 20}
		mockNoteRepo.On("FindByUserID", ctx, userID, 10, 20).Return([]*note.Note{}, nil).Once()

		_, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		mockNoteRepo.AssertExpectations(t)
	})
}

func TestNoteService_Copy(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	originalNoteID := int64(100)
	noteTypeID := int64(10)
	originalDeckID := int64(20)
	targetDeckID := int64(30)

	t.Run("Success with all options", func(t *testing.T) {
		// Setup original note
		originalNote, _ := note.NewBuilder().
			WithID(originalNoteID).
			WithUserID(userID).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Q", "Back":"A"}`).
			WithTags([]string{"tag1", "tag2"}).
			Build()

		// Setup note type
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithCardTypesJSON("[{}, {}]"). // 2 card types
			Build()

		// Setup deck
		d, _ := deck.NewBuilder().WithID(targetDeckID).WithUserID(userID).WithName("Target Deck").Build()

		// Setup original cards
		originalCard, _ := card.NewBuilder().
			WithNoteID(originalNoteID).
			WithDeckID(originalDeckID).
			Build()

		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(originalNote, nil).Once()
		mockDeckRepo.On("FindByID", mock.Anything, userID, targetDeckID).Return(d, nil).Once()
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockNoteRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*note.Note")).Return(nil).Run(func(args mock.Arguments) {
			n := args.Get(2).(*note.Note)
			n.SetID(200) // Set ID so card generation works
		}).Once()
		mockCardRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*card.Card")).Return(nil).Twice()

		result, err := service.Copy(ctx, userID, originalNoteID, &targetDeckID, true, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(200), result.GetID())
		assert.Equal(t, noteTypeID, result.GetNoteTypeID())
		assert.Equal(t, `{"Front":"Q", "Back":"A"}`, result.GetFieldsJSON())
		assert.Equal(t, []string{"tag1", "tag2"}, result.GetTags())
		assert.False(t, result.GetMarked()) // Should not inherit marked status
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockNoteTypeRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Success with custom deck", func(t *testing.T) {
		originalNote, _ := note.NewBuilder().
			WithID(originalNoteID).
			WithUserID(userID).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Q"}`).
			WithTags([]string{"tag1"}).
			Build()

		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithCardTypesJSON("[{}]"). // 1 card type
			Build()

		d, _ := deck.NewBuilder().WithID(targetDeckID).WithUserID(userID).WithName("Target Deck").Build()

		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(originalNote, nil).Once()
		mockDeckRepo.On("FindByID", mock.Anything, userID, targetDeckID).Return(d, nil).Once()
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockNoteRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*note.Note")).Return(nil).Run(func(args mock.Arguments) {
			n := args.Get(2).(*note.Note)
			n.SetID(201)
		}).Once()
		mockCardRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*card.Card")).Return(nil).Once()

		result, err := service.Copy(ctx, userID, originalNoteID, &targetDeckID, false, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.GetTags()) // Should not copy tags
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Success without tags", func(t *testing.T) {
		originalNote, _ := note.NewBuilder().
			WithID(originalNoteID).
			WithUserID(userID).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Q"}`).
			WithTags([]string{"tag1", "tag2"}).
			Build()

		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithCardTypesJSON("[{}]").
			Build()

		originalCard, _ := card.NewBuilder().
			WithNoteID(originalNoteID).
			WithDeckID(originalDeckID).
			Build()

		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(originalNote, nil).Once()
		mockCardRepo.On("FindByNoteID", mock.Anything, userID, originalNoteID).Return([]*card.Card{originalCard}, nil).Once()
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockNoteRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*note.Note")).Return(nil).Run(func(args mock.Arguments) {
			n := args.Get(2).(*note.Note)
			n.SetID(202)
		}).Once()
		mockCardRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*card.Card")).Return(nil).Once()

		result, err := service.Copy(ctx, userID, originalNoteID, nil, false, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.GetTags()) // Should not copy tags when copyTags is false
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Note not found", func(t *testing.T) {
		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(nil, nil).Once()

		result, err := service.Copy(ctx, userID, originalNoteID, nil, true, true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
		assert.Nil(t, result)
		mockNoteRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Deck not found", func(t *testing.T) {
		originalNote, _ := note.NewBuilder().
			WithID(originalNoteID).
			WithUserID(userID).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Q"}`).
			Build()

		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(originalNote, nil).Once()
		mockDeckRepo.On("FindByID", mock.Anything, userID, targetDeckID).Return(nil, nil).Once()

		result, err := service.Copy(ctx, userID, originalNoteID, &targetDeckID, true, true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deck not found")
		assert.Nil(t, result)
		mockNoteRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Note has no cards", func(t *testing.T) {
		originalNote, _ := note.NewBuilder().
			WithID(originalNoteID).
			WithUserID(userID).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Q"}`).
			Build()

		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(originalNote, nil).Once()
		mockCardRepo.On("FindByNoteID", mock.Anything, userID, originalNoteID).Return([]*card.Card{}, nil).Once()

		result, err := service.Copy(ctx, userID, originalNoteID, nil, true, true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note has no cards")
		assert.Nil(t, result)
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Cross-user isolation", func(t *testing.T) {
		otherUserID := int64(999)
		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(nil, nil).Once()

		result, err := service.Copy(ctx, userID, originalNoteID, nil, true, true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
		assert.Nil(t, result)
		mockNoteRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Generates new GUID automatically", func(t *testing.T) {
		// Setup original note with a specific GUID
		originalGUID, _ := valueobjects.NewGUID("550e8400-e29b-41d4-a716-446655440000")
		originalNote, _ := note.NewBuilder().
			WithID(originalNoteID).
			WithUserID(userID).
			WithGUID(originalGUID).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(`{"Front":"Q", "Back":"A"}`).
			WithTags([]string{"tag1"}).
			Build()

		// Setup note type
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithCardTypesJSON("[{}]"). // 1 card type
			Build()

		// Setup deck
		d, _ := deck.NewBuilder().WithID(targetDeckID).WithUserID(userID).WithName("Target Deck").Build()

		// Setup original cards
		originalCard, _ := card.NewBuilder().
			WithNoteID(originalNoteID).
			WithDeckID(originalDeckID).
			Build()

		mockTM.ExpectTransaction()
		mockNoteRepo.On("FindByID", mock.Anything, userID, originalNoteID).Return(originalNote, nil).Once()
		mockDeckRepo.On("FindByID", mock.Anything, userID, targetDeckID).Return(d, nil).Once()
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockNoteRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*note.Note")).Return(nil).Run(func(args mock.Arguments) {
			n := args.Get(2).(*note.Note)
			n.SetID(200) // Set ID so card generation works
			// Verify new GUID is generated and different from original
			newGUID := n.GetGUID()
			assert.False(t, newGUID.IsEmpty(), "New GUID should not be empty")
			assert.NotEmpty(t, newGUID.Value(), "New GUID should have a value")
			// Verify GUID format is valid UUID (RFC 4122)
			guidValue := newGUID.Value()
			assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, guidValue, "GUID should be in UUID format")
			// Verify new GUID is different from original
			assert.False(t, newGUID.Equals(originalGUID), "Copied note should have a different GUID from original")
		}).Once()
		mockCardRepo.On("Save", mock.Anything, userID, mock.AnythingOfType("*card.Card")).Return(nil).Once()

		result, err := service.Copy(ctx, userID, originalNoteID, &targetDeckID, true, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Verify GUID is present and different from original
		newGUID := result.GetGUID()
		assert.False(t, newGUID.IsEmpty(), "GUID should not be empty")
		assert.NotEmpty(t, newGUID.Value(), "GUID should have a value")
		assert.False(t, newGUID.Equals(originalGUID), "Copied note should have a different GUID from original")
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockDeckRepo.AssertExpectations(t)
		mockNoteTypeRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestNoteService_FindDuplicates(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	fieldName := "Front"
	noteTypeID := int64(10)

	t.Run("Success with note type filter", func(t *testing.T) {
		// Setup note type
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
			Build()

		// Setup duplicate groups
		groups := []*note.DuplicateGroup{
			{
				FieldValue: "Hello",
				Notes: []*note.DuplicateNoteInfo{
					{ID: 1, GUID: "guid1", DeckID: 20, CreatedAt: time.Now()},
					{ID: 2, GUID: "guid2", DeckID: 20, CreatedAt: time.Now()},
				},
			},
		}

		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		mockNoteRepo.On("FindDuplicatesByField", mock.Anything, userID, &noteTypeID, fieldName).Return(groups, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, fieldName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		assert.Len(t, result.Duplicates, 1)
		assert.Equal(t, "Hello", result.Duplicates[0].FieldValue)
		assert.Len(t, result.Duplicates[0].Notes, 2)
		mockNoteRepo.AssertExpectations(t)
		mockNoteTypeRepo.AssertExpectations(t)
	})

	t.Run("Success without note type filter", func(t *testing.T) {
		groups := []*note.DuplicateGroup{
			{
				FieldValue: "World",
				Notes: []*note.DuplicateNoteInfo{
					{ID: 3, GUID: "guid3", DeckID: 21, CreatedAt: time.Now()},
					{ID: 4, GUID: "guid4", DeckID: 21, CreatedAt: time.Now()},
				},
			},
		}

		mockNoteRepo.On("FindDuplicatesByField", mock.Anything, userID, (*int64)(nil), fieldName).Return(groups, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, nil, fieldName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Success no duplicates found", func(t *testing.T) {
		mockNoteRepo.On("FindDuplicatesByField", mock.Anything, userID, (*int64)(nil), fieldName).Return([]*note.DuplicateGroup{}, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, nil, fieldName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.Total)
		assert.Empty(t, result.Duplicates)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Empty field name", func(t *testing.T) {
		result, err := service.FindDuplicates(ctx, userID, nil, "")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.Total)
	})

	t.Run("Note type not found", func(t *testing.T) {
		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nil, ownership.ErrResourceNotFound).Once()

		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, fieldName)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note type not found")
		assert.Nil(t, result)
		mockNoteTypeRepo.AssertExpectations(t)
	})

	t.Run("Invalid field name", func(t *testing.T) {
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
			Build()

		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, "InvalidField")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field 'InvalidField' not found")
		assert.Nil(t, result)
		mockNoteTypeRepo.AssertExpectations(t)
	})

	t.Run("Success with automatic first field detection", func(t *testing.T) {
		// Setup note type with first field "Front"
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
			Build()

		// Setup duplicate groups
		groups := []*note.DuplicateGroup{
			{
				FieldValue: "Hello",
				Notes: []*note.DuplicateNoteInfo{
					{ID: 1, GUID: "guid1", DeckID: 20, CreatedAt: time.Now()},
					{ID: 2, GUID: "guid2", DeckID: 20, CreatedAt: time.Now()},
				},
			},
		}

		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		// Should use "Front" (first field) automatically
		mockNoteRepo.On("FindDuplicatesByField", mock.Anything, userID, &noteTypeID, "Front").Return(groups, nil).Once()

		// Call with empty fieldName - should automatically use first field
		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, "")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		assert.Len(t, result.Duplicates, 1)
		assert.Equal(t, "Hello", result.Duplicates[0].FieldValue)
		mockNoteRepo.AssertExpectations(t)
		mockNoteTypeRepo.AssertExpectations(t)
	})

	t.Run("Error when note type has no fields", func(t *testing.T) {
		// Setup note type with empty fields
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithFieldsJSON(`[]`).
			Build()

		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note type has no fields defined")
		assert.Nil(t, result)
		mockNoteTypeRepo.AssertExpectations(t)
	})

	t.Run("Error when note type fields JSON is invalid", func(t *testing.T) {
		// Setup note type with invalid fields JSON
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithFieldsJSON(`invalid json`).
			Build()

		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid note type fields JSON")
		assert.Nil(t, result)
		mockNoteTypeRepo.AssertExpectations(t)
	})

	t.Run("Error when first field has no name property", func(t *testing.T) {
		// Setup note type with first field missing name
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithFieldsJSON(`[{"ord":0},{"name":"Back"}]`).
			Build()

		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first field has no name property")
		assert.Nil(t, result)
		mockNoteTypeRepo.AssertExpectations(t)
	})

	t.Run("Backward compatibility - explicit field name still works", func(t *testing.T) {
		// Setup note type
		nt, _ := notetype.NewBuilder().
			WithID(noteTypeID).
			WithUserID(userID).
			WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
			Build()

		// Setup duplicate groups
		groups := []*note.DuplicateGroup{
			{
				FieldValue: "World",
				Notes: []*note.DuplicateNoteInfo{
					{ID: 3, GUID: "guid3", DeckID: 21, CreatedAt: time.Now()},
				},
			},
		}

		mockNoteTypeRepo.On("FindByID", mock.Anything, userID, noteTypeID).Return(nt, nil).Once()
		// Should use explicit field name "Back" instead of first field "Front"
		mockNoteRepo.On("FindDuplicatesByField", mock.Anything, userID, &noteTypeID, "Back").Return(groups, nil).Once()

		result, err := service.FindDuplicates(ctx, userID, &noteTypeID, "Back")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		mockNoteRepo.AssertExpectations(t)
		mockNoteTypeRepo.AssertExpectations(t)
	})
}

func TestNoteService_FindDuplicatesByGUID(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success with duplicates found", func(t *testing.T) {
		// Setup duplicate groups with same GUID
		groups := []*note.DuplicateGroup{
			{
				FieldValue: "550e8400-e29b-41d4-a716-446655440000",
				Notes: []*note.DuplicateNoteInfo{
					{ID: 1, GUID: "550e8400-e29b-41d4-a716-446655440000", DeckID: 20, CreatedAt: time.Now()},
					{ID: 2, GUID: "550e8400-e29b-41d4-a716-446655440000", DeckID: 21, CreatedAt: time.Now()},
				},
			},
		}

		mockNoteRepo.On("FindDuplicatesByGUID", mock.Anything, userID).Return(groups, nil).Once()

		result, err := service.FindDuplicatesByGUID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		assert.Len(t, result.Duplicates, 1)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result.Duplicates[0].FieldValue)
		assert.Len(t, result.Duplicates[0].Notes, 2)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Success no duplicates found", func(t *testing.T) {
		mockNoteRepo.On("FindDuplicatesByGUID", mock.Anything, userID).Return([]*note.DuplicateGroup{}, nil).Once()

		result, err := service.FindDuplicatesByGUID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.Total)
		assert.Empty(t, result.Duplicates)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		expectedErr := fmt.Errorf("database error")
		mockNoteRepo.On("FindDuplicatesByGUID", mock.Anything, userID).Return(nil, expectedErr).Once()

		result, err := service.FindDuplicatesByGUID(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Multiple duplicate groups", func(t *testing.T) {
		groups := []*note.DuplicateGroup{
			{
				FieldValue: "guid1",
				Notes: []*note.DuplicateNoteInfo{
					{ID: 1, GUID: "guid1", DeckID: 20, CreatedAt: time.Now()},
					{ID: 2, GUID: "guid1", DeckID: 21, CreatedAt: time.Now()},
				},
			},
			{
				FieldValue: "guid2",
				Notes: []*note.DuplicateNoteInfo{
					{ID: 3, GUID: "guid2", DeckID: 22, CreatedAt: time.Now()},
					{ID: 4, GUID: "guid2", DeckID: 23, CreatedAt: time.Now()},
				},
			},
		}

		mockNoteRepo.On("FindDuplicatesByGUID", mock.Anything, userID).Return(groups, nil).Once()

		result, err := service.FindDuplicatesByGUID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Total)
		assert.Len(t, result.Duplicates, 2)
		mockNoteRepo.AssertExpectations(t)
	})
}

func TestNoteService_FindByID(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	noteID := int64(100)

	t.Run("Success", func(t *testing.T) {
		expectedNote := &note.Note{}
		expectedNote.SetID(noteID)
		mockNoteRepo.On("FindByID", ctx, userID, noteID).Return(expectedNote, nil).Once()

		result, err := service.FindByID(ctx, userID, noteID)

		assert.NoError(t, err)
		assert.Equal(t, expectedNote, result)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockNoteRepo.On("FindByID", ctx, userID, noteID).Return(nil, nil).Once()

		result, err := service.FindByID(ctx, userID, noteID)

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockNoteRepo.AssertExpectations(t)
	})
}

func TestNoteService_Update(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	noteID := int64(100)

	t.Run("Success", func(t *testing.T) {
		existing := &note.Note{}
		existing.SetID(noteID)
		fields := "{\"Front\":\"New Q\"}"
		tags := []string{"new-tag"}

		mockNoteRepo.On("FindByID", ctx, userID, noteID).Return(existing, nil).Once()
		mockNoteRepo.On("Update", ctx, userID, noteID, existing).Return(nil).Once()

		result, err := service.Update(ctx, userID, noteID, fields, tags)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, fields, result.GetFieldsJSON())
		assert.Equal(t, tags, result.GetTags())
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockNoteRepo.On("FindByID", ctx, userID, noteID).Return(nil, nil).Once()

		result, err := service.Update(ctx, userID, noteID, "{}", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
		assert.Nil(t, result)
	})
}

func TestNoteService_AddTag(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	noteID := int64(100)

	t.Run("Success", func(t *testing.T) {
		existing := &note.Note{}
		existing.SetID(noteID)
		tag := "new-tag"

		mockNoteRepo.On("FindByID", ctx, userID, noteID).Return(existing, nil).Once()
		mockNoteRepo.On("Update", ctx, userID, noteID, existing).Return(nil).Once()

		err := service.AddTag(ctx, userID, noteID, tag)

		assert.NoError(t, err)
		assert.True(t, existing.HasTag(tag))
		mockNoteRepo.AssertExpectations(t)
	})
}

func TestNoteService_RemoveTag(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	noteID := int64(100)

	t.Run("Success", func(t *testing.T) {
		existing := &note.Note{}
		existing.SetID(noteID)
		existing.SetTags([]string{"tag1", "tag2"})
		tag := "tag1"

		mockNoteRepo.On("FindByID", ctx, userID, noteID).Return(existing, nil).Once()
		mockNoteRepo.On("Update", ctx, userID, noteID, existing).Return(nil).Once()

		err := service.RemoveTag(ctx, userID, noteID, tag)

		assert.NoError(t, err)
		assert.False(t, existing.HasTag(tag))
		mockNoteRepo.AssertExpectations(t)
	})
}

func TestNoteService_Delete(t *testing.T) {
	mockNoteRepo := new(MockNoteRepository)
	mockCardRepo := new(MockCardRepository)
	mockNoteTypeRepo := new(MockNoteTypeRepository)
	mockDeckRepo := new(MockDeckRepository)
	mockTM := new(MockTransactionManager)
	service := noteSvc.NewNoteService(mockNoteRepo, mockCardRepo, mockNoteTypeRepo, mockDeckRepo, mockTM)
	ctx := context.Background()
	userID := int64(1)
	noteID := int64(100)

	t.Run("Success", func(t *testing.T) {
		cards := []*card.Card{
			{ /* card 1 */ },
			{ /* card 2 */ },
		}

		mockTM.ExpectTransaction()
		mockNoteRepo.On("Delete", mock.Anything, userID, noteID).Return(nil).Once()
		mockCardRepo.On("FindByNoteID", mock.Anything, userID, noteID).Return(cards, nil).Once()
		mockCardRepo.On("Delete", mock.Anything, userID, mock.Anything).Return(nil).Twice()

		err := service.Delete(ctx, userID, noteID)

		assert.NoError(t, err)
		mockNoteRepo.AssertExpectations(t)
		mockCardRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

