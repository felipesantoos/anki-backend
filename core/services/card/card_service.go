package card

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/database"
)

// CardService implements ICardService
type CardService struct {
	cardRepo        secondary.ICardRepository
	noteService     primary.INoteService
	deckService     primary.IDeckService
	noteTypeService primary.INoteTypeService
	reviewService   primary.IReviewService
	tm              database.TransactionManager
}

// NewCardService creates a new CardService instance
func NewCardService(
	cardRepo secondary.ICardRepository,
	noteService primary.INoteService,
	deckService primary.IDeckService,
	noteTypeService primary.INoteTypeService,
	reviewService primary.IReviewService,
	tm database.TransactionManager,
) primary.ICardService {
	return &CardService{
		cardRepo:        cardRepo,
		noteService:     noteService,
		deckService:     deckService,
		noteTypeService: noteTypeService,
		reviewService:   reviewService,
		tm:              tm,
	}
}

// FindByID finds a card by ID
func (s *CardService) FindByID(ctx context.Context, userID int64, id int64) (*card.Card, error) {
	return s.cardRepo.FindByID(ctx, userID, id)
}

// FindByDeckID finds all cards in a deck
func (s *CardService) FindByDeckID(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error) {
	return s.cardRepo.FindByDeckID(ctx, userID, deckID)
}

// Update updates an existing card
func (s *CardService) Update(ctx context.Context, userID int64, cardEntity *card.Card) error {
	cardEntity.SetUpdatedAt(time.Now())
	return s.cardRepo.Update(ctx, userID, cardEntity.GetID(), cardEntity)
}

// Delete deletes a card
func (s *CardService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.cardRepo.Delete(ctx, userID, id)
}

// Suspend suspends a card
func (s *CardService) Suspend(ctx context.Context, userID int64, id int64) error {
	c, err := s.cardRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("card not found")
	}

	c.Suspend()
	return s.cardRepo.Update(ctx, userID, id, c)
}

// Unsuspend unsuspends a card
func (s *CardService) Unsuspend(ctx context.Context, userID int64, id int64) error {
	c, err := s.cardRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("card not found")
	}

	c.Unsuspend()
	return s.cardRepo.Update(ctx, userID, id, c)
}

// Bury buries a card
func (s *CardService) Bury(ctx context.Context, userID int64, id int64) error {
	c, err := s.cardRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("card not found")
	}

	c.Bury()
	return s.cardRepo.Update(ctx, userID, id, c)
}

// Unbury unburies a card
func (s *CardService) Unbury(ctx context.Context, userID int64, id int64) error {
	c, err := s.cardRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("card not found")
	}

	c.Unbury()
	return s.cardRepo.Update(ctx, userID, id, c)
}

// SetFlag sets a colored flag on a card
func (s *CardService) SetFlag(ctx context.Context, userID int64, id int64, flag int) error {
	c, err := s.cardRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("card not found")
	}

	if err := c.SetFlag(flag); err != nil {
		return err
	}
	return s.cardRepo.Update(ctx, userID, id, c)
}

// FindDueCards finds cards that are due for review in a deck
func (s *CardService) FindDueCards(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error) {
	now := time.Now().Unix() * 1000 // Current time in milliseconds
	return s.cardRepo.FindDueCards(ctx, userID, deckID, now)
}

// CountByDeckAndState counts cards with a specific state in a deck
func (s *CardService) CountByDeckAndState(ctx context.Context, userID int64, deckID int64, state string) (int, error) {
	cardState := valueobjects.CardState(state)
	if !cardState.IsValid() {
		return 0, fmt.Errorf("invalid card state: %s", state)
	}
	return s.cardRepo.CountByDeckAndState(ctx, userID, deckID, cardState)
}

// FindAll finds cards for a user based on filters and pagination
func (s *CardService) FindAll(ctx context.Context, userID int64, filters card.CardFilters) ([]*card.Card, int, error) {
	// Validate state filter if provided
	if filters.State != nil {
		cardState := valueobjects.CardState(*filters.State)
		if !cardState.IsValid() {
			return nil, 0, fmt.Errorf("invalid card state: %s", *filters.State)
		}
		stateStr := cardState.String()
		filters.State = &stateStr
	}

	// Validate flag filter if provided
	if filters.Flag != nil {
		if *filters.Flag < 0 || *filters.Flag > 7 {
			return nil, 0, fmt.Errorf("flag must be between 0 and 7")
		}
	}

	// Apply defaults for pagination
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	return s.cardRepo.FindAll(ctx, userID, filters)
}

// GetInfo returns detailed card information including note data, deck/note type names, and review history
func (s *CardService) GetInfo(ctx context.Context, userID int64, cardID int64) (*card.CardInfo, error) {
	// Get card (validates ownership)
	cardEntity, err := s.cardRepo.FindByID(ctx, userID, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to find card: %w", err)
	}

	// Get note
	noteEntity, err := s.noteService.FindByID(ctx, userID, cardEntity.GetNoteID())
	if err != nil {
		return nil, fmt.Errorf("failed to find note: %w", err)
	}

	// Get deck
	deckEntity, err := s.deckService.FindByID(ctx, userID, cardEntity.GetDeckID())
	if err != nil {
		return nil, fmt.Errorf("failed to find deck: %w", err)
	}

	// Get note type
	noteTypeEntity, err := s.noteTypeService.FindByID(ctx, userID, noteEntity.GetNoteTypeID())
	if err != nil {
		return nil, fmt.Errorf("failed to find note type: %w", err)
	}

	// Get reviews
	reviews, err := s.reviewService.FindByCardID(ctx, userID, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews: %w", err)
	}

	// Parse note fields JSON
	var fields map[string]interface{}
	if noteEntity.GetFieldsJSON() != "" {
		if err := json.Unmarshal([]byte(noteEntity.GetFieldsJSON()), &fields); err != nil {
			return nil, fmt.Errorf("failed to parse note fields: %w", err)
		}
	} else {
		fields = make(map[string]interface{})
	}

	// Calculate statistics from reviews
	var firstReview *time.Time
	var lastReview *time.Time
	totalReviews := len(reviews)
	easeHistory := make([]int, 0, totalReviews)
	intervalHistory := make([]int, 0, totalReviews)
	reviewHistory := make([]*card.ReviewInfo, 0, totalReviews)

	if totalReviews > 0 {
		// Reviews are ordered by createdAt DESC from repository, so first is last and last is first
		// We need to reverse for chronological order (oldest first)
		firstCreatedAt := reviews[totalReviews-1].GetCreatedAt()
		lastCreatedAt := reviews[0].GetCreatedAt()
		firstReview = &firstCreatedAt
		lastReview = &lastCreatedAt

		// Build histories in chronological order (oldest first)
		for i := totalReviews - 1; i >= 0; i-- {
			r := reviews[i]
			easeHistory = append(easeHistory, r.GetEase())
			intervalHistory = append(intervalHistory, r.GetInterval())
			reviewHistory = append(reviewHistory, &card.ReviewInfo{
				Rating:    r.GetRating(),
				Interval:  r.GetInterval(),
				Ease:      r.GetEase(),
				TimeMs:    r.GetTimeMs(),
				Type:      r.GetType().String(),
				CreatedAt: r.GetCreatedAt(),
			})
		}
	}

	return &card.CardInfo{
		CardID:          cardEntity.GetID(),
		NoteID:          cardEntity.GetNoteID(),
		DeckName:        deckEntity.GetName(),
		NoteTypeName:    noteTypeEntity.GetName(),
		Fields:          fields,
		Tags:            noteEntity.GetTags(),
		CreatedAt:       cardEntity.GetCreatedAt(),
		FirstReview:     firstReview,
		LastReview:      lastReview,
		TotalReviews:    totalReviews,
		EaseHistory:     easeHistory,
		IntervalHistory: intervalHistory,
		ReviewHistory:   reviewHistory,
	}, nil
}

// Reset resets a card (type can be "new" or "forget")
func (s *CardService) Reset(ctx context.Context, userID int64, id int64, resetType string) error {
	return s.tm.WithTransaction(ctx, func(txCtx context.Context) error {
		c, err := s.cardRepo.FindByID(txCtx, userID, id)
		if err != nil {
			return err
		}
		if c == nil {
			return fmt.Errorf("card not found")
		}

		if resetType == "forget" {
			c.Forget()
			if err := s.reviewService.DeleteByCardID(txCtx, userID, id); err != nil {
				return fmt.Errorf("failed to delete reviews: %w", err)
			}
		} else {
			c.Reset(true, true)
		}

		return s.cardRepo.Update(txCtx, userID, id, c)
	})
}

// SetDueDate manually sets the due date for a card
func (s *CardService) SetDueDate(ctx context.Context, userID int64, id int64, due int64) error {
	c, err := s.cardRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("card not found")
	}

	c.SetDue(due)
	return s.cardRepo.Update(ctx, userID, id, c)
}
