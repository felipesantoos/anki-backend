package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/review"
)

// IReviewService defines the interface for card review operations
type IReviewService interface {
	// Create records a new review for a card and updates the card's scheduling state
	Create(ctx context.Context, userID int64, cardID int64, rating int, timeMs int) (*review.Review, error)

	// FindByID finds a review by ID
	FindByID(ctx context.Context, userID int64, id int64) (*review.Review, error)

	// FindByCardID finds all reviews for a card
	FindByCardID(ctx context.Context, userID int64, cardID int64) ([]*review.Review, error)
}

