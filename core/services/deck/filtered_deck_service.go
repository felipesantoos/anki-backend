package deck

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// FilteredDeckService implements IFilteredDeckService
type FilteredDeckService struct {
	repo secondary.IFilteredDeckRepository
}

// NewFilteredDeckService creates a new FilteredDeckService instance
func NewFilteredDeckService(repo secondary.IFilteredDeckRepository) primary.IFilteredDeckService {
	return &FilteredDeckService{
		repo: repo,
	}
}

// Create creates a new filtered deck
func (s *FilteredDeckService) Create(ctx context.Context, userID int64, name string, searchFilter string, limit int, orderBy string, reschedule bool) (*filtereddeck.FilteredDeck, error) {
	now := time.Now()
	fd, err := filtereddeck.NewBuilder().
		WithUserID(userID).
		WithName(name).
		WithSearchFilter(searchFilter).
		WithLimitCards(limit).
		WithOrderBy(orderBy).
		WithReschedule(reschedule).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, fd); err != nil {
		return nil, err
	}

	return fd, nil
}

// FindByUserID finds all filtered decks for a user
func (s *FilteredDeckService) FindByUserID(ctx context.Context, userID int64) ([]*filtereddeck.FilteredDeck, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// Update updates an existing filtered deck
func (s *FilteredDeckService) Update(ctx context.Context, userID int64, id int64, name string, searchFilter string, limit int, orderBy string, reschedule bool) (*filtereddeck.FilteredDeck, error) {
	existing, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("filtered deck not found")
	}

	existing.SetName(name)
	existing.SetSearchFilter(searchFilter)
	existing.SetLimitCards(limit)
	existing.SetOrderBy(orderBy)
	existing.SetReschedule(reschedule)
	existing.SetUpdatedAt(time.Now())

	if err := s.repo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a filtered deck
func (s *FilteredDeckService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

