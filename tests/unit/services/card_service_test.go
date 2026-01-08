package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	cardSvc "github.com/felipesantos/anki-backend/core/services/card"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCardService_Suspend(t *testing.T) {
	mockRepo := new(MockCardRepository)
	mockNoteSvc := new(MockNoteService)
	mockDeckSvc := new(MockDeckService)
	mockNoteTypeSvc := new(MockNoteTypeService)
	mockReviewSvc := new(MockReviewService)
	mockTM := new(MockTransactionManager)
	service := cardSvc.NewCardService(mockRepo, mockNoteSvc, mockDeckSvc, mockNoteTypeSvc, mockReviewSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)
	cardID := int64(100)

	t.Run("Success", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithNoteID(1).WithDeckID(1).Build()
		assert.False(t, c.GetSuspended())

		mockRepo.On("FindByID", ctx, userID, cardID).Return(c, nil).Once()
		mockRepo.On("Update", ctx, userID, cardID, mock.Anything).Return(nil).Once()

		err := service.Suspend(ctx, userID, cardID)

		assert.NoError(t, err)
		assert.True(t, c.GetSuspended())
		mockRepo.AssertExpectations(t)
	})
}

func TestCardService_SetFlag(t *testing.T) {
	mockRepo := new(MockCardRepository)
	mockNoteSvc := new(MockNoteService)
	mockDeckSvc := new(MockDeckService)
	mockNoteTypeSvc := new(MockNoteTypeService)
	mockReviewSvc := new(MockReviewService)
	mockTM := new(MockTransactionManager)
	service := cardSvc.NewCardService(mockRepo, mockNoteSvc, mockDeckSvc, mockNoteTypeSvc, mockReviewSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)
	cardID := int64(100)

	t.Run("Success", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithNoteID(1).WithDeckID(1).Build()
		flag := 3

		mockRepo.On("FindByID", ctx, userID, cardID).Return(c, nil).Once()
		mockRepo.On("Update", ctx, userID, cardID, mock.Anything).Return(nil).Once()

		err := service.SetFlag(ctx, userID, cardID, flag)

		assert.NoError(t, err)
		assert.Equal(t, flag, c.GetFlag())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Flag", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithNoteID(1).WithDeckID(1).Build()
		flag := 9 // Invalid

		mockRepo.On("FindByID", ctx, userID, cardID).Return(c, nil).Once()

		err := service.SetFlag(ctx, userID, cardID, flag)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCardService_CountByDeckAndState(t *testing.T) {
	mockRepo := new(MockCardRepository)
	mockNoteSvc := new(MockNoteService)
	mockDeckSvc := new(MockDeckService)
	mockNoteTypeSvc := new(MockNoteTypeService)
	mockReviewSvc := new(MockReviewService)
	mockTM := new(MockTransactionManager)
	service := cardSvc.NewCardService(mockRepo, mockNoteSvc, mockDeckSvc, mockNoteTypeSvc, mockReviewSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)
	deckID := int64(10)

	t.Run("Success New", func(t *testing.T) {
		mockRepo.On("CountByDeckAndState", ctx, userID, deckID, valueobjects.CardStateNew).Return(5, nil).Once()

		count, err := service.CountByDeckAndState(ctx, userID, deckID, "new")

		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid State", func(t *testing.T) {
		count, err := service.CountByDeckAndState(ctx, userID, deckID, "invalid")

		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "invalid card state")
	})
}

func TestCardService_FindAll(t *testing.T) {
	mockRepo := new(MockCardRepository)
	mockNoteSvc := new(MockNoteService)
	mockDeckSvc := new(MockDeckService)
	mockNoteTypeSvc := new(MockNoteTypeService)
	mockReviewSvc := new(MockReviewService)
	mockTM := new(MockTransactionManager)
	service := cardSvc.NewCardService(mockRepo, mockNoteSvc, mockDeckSvc, mockNoteTypeSvc, mockReviewSvc, mockTM)
	ctx := context.Background()
	userID := int64(1)

	t.Run("Success without filters", func(t *testing.T) {
		cards := []*card.Card{
			func() *card.Card { c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).Build(); return c }(),
			func() *card.Card { c, _ := card.NewBuilder().WithID(2).WithNoteID(2).WithDeckID(1).Build(); return c }(),
		}
		total := 2

		filters := card.CardFilters{
			Limit:  20,
			Offset: 0,
		}

		mockRepo.On("FindAll", ctx, userID, filters).Return(cards, total, nil).Once()

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, cards, result)
		assert.Equal(t, total, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with deck filter", func(t *testing.T) {
		deckID := int64(10)
		cards := []*card.Card{
			func() *card.Card { c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(deckID).Build(); return c }(),
		}
		total := 1

		filters := card.CardFilters{
			DeckID: &deckID,
			Limit:  20,
			Offset: 0,
		}

		mockRepo.On("FindAll", ctx, userID, filters).Return(cards, total, nil).Once()

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, cards, result)
		assert.Equal(t, total, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with state filter", func(t *testing.T) {
		state := "new"
		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).WithState(valueobjects.CardStateNew).Build()
				return c
			}(),
		}
		total := 1

		filters := card.CardFilters{
			State:  &state,
			Limit:  20,
			Offset: 0,
		}

		mockRepo.On("FindAll", ctx, userID, mock.Anything).Return(cards, total, nil).Once()

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, cards, result)
		assert.Equal(t, total, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with pagination", func(t *testing.T) {
		cards := []*card.Card{
			func() *card.Card { c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(1).Build(); return c }(),
		}
		total := 25

		filters := card.CardFilters{
			Limit:  10,
			Offset: 10,
		}

		mockRepo.On("FindAll", ctx, userID, filters).Return(cards, total, nil).Once()

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, cards, result)
		assert.Equal(t, total, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Applies default pagination", func(t *testing.T) {
		cards := []*card.Card{}
		total := 0

		filters := card.CardFilters{
			Limit:  0, // Should default to 20
			Offset: -1, // Should default to 0
		}

		expectedFilters := card.CardFilters{
			Limit:  20,
			Offset: 0,
		}

		mockRepo.On("FindAll", ctx, userID, expectedFilters).Return(cards, total, nil).Once()

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, cards, result)
		assert.Equal(t, total, count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid state filter", func(t *testing.T) {
		invalidState := "invalid"
		filters := card.CardFilters{
			State:  &invalidState,
			Limit:  20,
			Offset: 0,
		}

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "invalid card state")
	})

	t.Run("Invalid flag filter", func(t *testing.T) {
		invalidFlag := 10
		filters := card.CardFilters{
			Flag:   &invalidFlag,
			Limit:  20,
			Offset: 0,
		}

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "flag must be between 0 and 7")
	})

	t.Run("Success with all filters", func(t *testing.T) {
		deckID := int64(10)
		state := "review"
		flag := 3
		suspended := false
		buried := false
		cards := []*card.Card{
			func() *card.Card {
				c, _ := card.NewBuilder().WithID(1).WithNoteID(1).WithDeckID(deckID).WithState(valueobjects.CardStateReview).Build()
				c.SetFlag(flag)
				return c
			}(),
		}
		total := 1

		filters := card.CardFilters{
			DeckID:    &deckID,
			State:     &state,
			Flag:      &flag,
			Suspended: &suspended,
			Buried:    &buried,
			Limit:     20,
			Offset:    0,
		}

		mockRepo.On("FindAll", ctx, userID, mock.Anything).Return(cards, total, nil).Once()

		result, count, err := service.FindAll(ctx, userID, filters)

		assert.NoError(t, err)
		assert.Equal(t, cards, result)
		assert.Equal(t, total, count)
		mockRepo.AssertExpectations(t)
	})
}

func TestCardService_Reset(t *testing.T) {
	mockRepo := new(MockCardRepository)
	mockNoteSvc := new(MockNoteService)
	mockDeckSvc := new(MockDeckService)
	mockNoteTypeSvc := new(MockNoteTypeService)
	mockReviewSvc := new(MockReviewService)
	mockTM := new(MockTransactionManager)
	service := cardSvc.NewCardService(mockRepo, mockNoteSvc, mockDeckSvc, mockNoteTypeSvc, mockReviewSvc, mockTM)

	ctx := context.Background()
	userID := int64(1)
	cardID := int64(123)

	t.Run("Success - Reset type new", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithDeckID(1).WithState(valueobjects.CardStateReview).Build()
		c.SetReps(10)
		c.SetLapses(2)

		mockTM.On("WithTransaction", ctx, mock.Anything).Return(nil).Once()
		mockRepo.On("FindByID", mock.Anything, userID, cardID).Return(c, nil).Once()
		mockRepo.On("Update", mock.Anything, userID, cardID, mock.Anything).Return(nil).Once()

		err := service.Reset(ctx, userID, cardID, "new")
		assert.NoError(t, err)
		assert.Equal(t, valueobjects.CardStateNew, c.GetState())
		assert.Equal(t, 0, c.GetReps())
		assert.Equal(t, 0, c.GetLapses())
		mockTM.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Reset type forget", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithDeckID(1).WithState(valueobjects.CardStateReview).Build()

		mockTM.On("WithTransaction", ctx, mock.Anything).Return(nil).Once()
		mockRepo.On("FindByID", mock.Anything, userID, cardID).Return(c, nil).Once()
		mockReviewSvc.On("DeleteByCardID", mock.Anything, userID, cardID).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, userID, cardID, mock.Anything).Return(nil).Once()

		err := service.Reset(ctx, userID, cardID, "forget")
		assert.NoError(t, err)
		assert.Equal(t, valueobjects.CardStateNew, c.GetState())
		mockTM.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockReviewSvc.AssertExpectations(t)
	})

	t.Run("Card not found", func(t *testing.T) {
		mockTM.On("WithTransaction", ctx, mock.Anything).Return(fmt.Errorf("card not found")).Once()
		mockRepo.On("FindByID", mock.Anything, userID, cardID).Return(nil, nil).Once()

		err := service.Reset(ctx, userID, cardID, "new")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card not found")
	})
}

func TestCardService_SetDueDate(t *testing.T) {
	mockRepo := new(MockCardRepository)
	mockNoteSvc := new(MockNoteService)
	mockDeckSvc := new(MockDeckService)
	mockNoteTypeSvc := new(MockNoteTypeService)
	mockReviewSvc := new(MockReviewService)
	mockTM := new(MockTransactionManager)
	service := cardSvc.NewCardService(mockRepo, mockNoteSvc, mockDeckSvc, mockNoteTypeSvc, mockReviewSvc, mockTM)

	ctx := context.Background()
	userID := int64(1)
	cardID := int64(123)
	due := int64(1705324200000)

	t.Run("Success", func(t *testing.T) {
		c, _ := card.NewBuilder().WithID(cardID).WithDeckID(1).WithDue(0).Build()

		mockRepo.On("FindByID", ctx, userID, cardID).Return(c, nil).Once()
		mockRepo.On("Update", ctx, userID, cardID, mock.Anything).Return(nil).Once()

		err := service.SetDueDate(ctx, userID, cardID, due)
		assert.NoError(t, err)
		assert.Equal(t, due, c.GetDue())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Card not found", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, userID, cardID).Return(nil, nil).Once()

		err := service.SetDueDate(ctx, userID, cardID, due)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card not found")
		mockRepo.AssertExpectations(t)
	})
}
