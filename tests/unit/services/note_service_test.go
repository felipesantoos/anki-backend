package services

import (
	"context"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	noteSvc "github.com/felipesantos/anki-backend/core/services/note"
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

