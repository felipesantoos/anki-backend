package card

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// CardService implements ICardService
type CardService struct {
	cardRepo secondary.ICardRepository
}

// NewCardService creates a new CardService instance
func NewCardService(cardRepo secondary.ICardRepository) primary.ICardService {
	return &CardService{
		cardRepo: cardRepo,
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

