package shareddeckrating

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// SharedDeckRatingService implements ISharedDeckRatingService
type SharedDeckRatingService struct {
	repo secondary.ISharedDeckRatingRepository
}

// NewSharedDeckRatingService creates a new SharedDeckRatingService instance
func NewSharedDeckRatingService(repo secondary.ISharedDeckRatingRepository) primary.ISharedDeckRatingService {
	return &SharedDeckRatingService{
		repo: repo,
	}
}

// Create records a new rating for a shared deck
func (s *SharedDeckRatingService) Create(ctx context.Context, userID int64, sharedDeckID int64, rating int, comment *string) (*shareddeckrating.SharedDeckRating, error) {
	// Check if user already rated this deck
	existing, err := s.repo.FindByUserIDAndSharedDeckID(ctx, userID, sharedDeckID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user already rated this shared deck")
	}

	now := time.Now()
	ratingVO, err := valueobjects.NewSharedDeckRating(rating)
	if err != nil {
		return nil, err
	}

	r, err := shareddeckrating.NewBuilder().
		WithUserID(userID).
		WithSharedDeckID(sharedDeckID).
		WithRating(ratingVO).
		WithComment(comment).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, r); err != nil {
		return nil, err
	}

	return r, nil
}

// FindBySharedDeckID finds all ratings for a shared deck
func (s *SharedDeckRatingService) FindBySharedDeckID(ctx context.Context, sharedDeckID int64) ([]*shareddeckrating.SharedDeckRating, error) {
	// Using default values for pagination
	return s.repo.FindBySharedDeckID(ctx, sharedDeckID, 0, 100)
}

// Update updates an existing rating
func (s *SharedDeckRatingService) Update(ctx context.Context, userID int64, sharedDeckID int64, rating int, comment *string) (*shareddeckrating.SharedDeckRating, error) {
	existing, err := s.repo.FindByUserIDAndSharedDeckID(ctx, userID, sharedDeckID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("rating not found")
	}

	ratingVO, err := valueobjects.NewSharedDeckRating(rating)
	if err != nil {
		return nil, err
	}

	existing.SetRating(ratingVO)
	existing.SetComment(comment)
	existing.SetUpdatedAt(time.Now())

	if err := s.repo.Update(ctx, userID, existing.GetID(), existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes a rating
func (s *SharedDeckRatingService) Delete(ctx context.Context, userID int64, sharedDeckID int64) error {
	existing, err := s.repo.FindByUserIDAndSharedDeckID(ctx, userID, sharedDeckID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("rating not found")
	}

	return s.repo.Delete(ctx, userID, existing.GetID())
}

