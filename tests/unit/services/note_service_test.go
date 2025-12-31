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

