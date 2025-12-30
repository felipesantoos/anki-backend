package savedsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// SavedSearchService implements ISavedSearchService
type SavedSearchService struct {
	repo secondary.ISavedSearchRepository
}

// NewSavedSearchService creates a new SavedSearchService instance
func NewSavedSearchService(repo secondary.ISavedSearchRepository) primary.ISavedSearchService {
	return &SavedSearchService{
		repo: repo,
	}
}

// Create creates a new saved search
func (s *SavedSearchService) Create(ctx context.Context, userID int64, name string, query string) (*savedsearch.SavedSearch, error) {
	exists, err := s.repo.FindByName(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, fmt.Errorf("saved search with name %s already exists", name)
	}

	now := time.Now()
	ss, err := savedsearch.NewBuilder().
		WithUserID(userID).
		WithName(name).
		WithSearchQuery(query).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, ss); err != nil {
		return nil, err
	}

	return ss, nil
}

// FindByUserID finds all saved searches for a user
func (s *SavedSearchService) FindByUserID(ctx context.Context, userID int64) ([]*savedsearch.SavedSearch, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// Update updates an existing saved search
func (s *SavedSearchService) Update(ctx context.Context, userID int64, id int64, name string, query string) (*savedsearch.SavedSearch, error) {
	existing, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("saved search not found")
	}

	if existing.GetName() != name {
		conflict, err := s.repo.FindByName(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if conflict != nil {
			return nil, fmt.Errorf("saved search with name %s already exists", name)
		}
	}

	existing.SetName(name)
	existing.SetSearchQuery(query)
	existing.SetUpdatedAt(time.Now())

	if err := s.repo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a saved search
func (s *SavedSearchService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

