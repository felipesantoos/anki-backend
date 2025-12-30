package review

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/database"
)

// ReviewService implements IReviewService
type ReviewService struct {
	reviewRepo secondary.IReviewRepository
	cardRepo   secondary.ICardRepository
	tm         database.TransactionManager
}

// NewReviewService creates a new ReviewService instance
func NewReviewService(
	reviewRepo secondary.IReviewRepository,
	cardRepo secondary.ICardRepository,
	tm database.TransactionManager,
) primary.IReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		cardRepo:   cardRepo,
		tm:         tm,
	}
}

// Create records a new review for a card and updates the card's scheduling state
func (s *ReviewService) Create(ctx context.Context, userID int64, cardID int64, rating int, timeMs int) (*review.Review, error) {
	var reviewEntity *review.Review

	err := s.tm.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Find and validate card
		c, err := s.cardRepo.FindByID(txCtx, userID, cardID)
		if err != nil {
			return err
		}
		if c == nil {
			return fmt.Errorf("card not found")
		}

		// 2. Map current state to review type
		var reviewType valueobjects.ReviewType
		switch c.GetState() {
		case valueobjects.CardStateLearn:
			reviewType = valueobjects.ReviewTypeLearn
		case valueobjects.CardStateReview:
			reviewType = valueobjects.ReviewTypeReview
		case valueobjects.CardStateRelearn:
			reviewType = valueobjects.ReviewTypeRelearn
		default:
			reviewType = valueobjects.ReviewTypeLearn
		}

		// 3. Update card state (simplified scheduling logic)
		now := time.Now()
		c.SetLastReviewAt(&now)
		c.SetReps(c.GetReps() + 1)
		if rating == 1 { // Again
			c.SetLapses(c.GetLapses() + 1)
			c.SetState(valueobjects.CardStateLearn) // Move back to learning if forgotten
		} else if c.GetState() == valueobjects.CardStateNew {
			c.SetState(valueobjects.CardStateLearn) // Move from new to learn
		}
		
		c.SetUpdatedAt(now)

		if err := s.cardRepo.Update(txCtx, userID, cardID, c); err != nil {
			return err
		}

		// 4. Create review record
		reviewEntity, err = review.NewBuilder().
			WithCardID(cardID).
			WithRating(rating).
			WithInterval(c.GetInterval()).
			WithEase(c.GetEase()).
			WithTimeMs(timeMs).
			WithType(reviewType).
			WithCreatedAt(now).
			Build()
		if err != nil {
			return err
		}

		if err := s.reviewRepo.Save(txCtx, userID, reviewEntity); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return reviewEntity, nil
}

// FindByID finds a review by ID
func (s *ReviewService) FindByID(ctx context.Context, userID int64, id int64) (*review.Review, error) {
	return s.reviewRepo.FindByID(ctx, userID, id)
}

// FindByCardID finds all reviews for a card
func (s *ReviewService) FindByCardID(ctx context.Context, userID int64, cardID int64) ([]*review.Review, error) {
	return s.reviewRepo.FindByCardID(ctx, userID, cardID)
}

