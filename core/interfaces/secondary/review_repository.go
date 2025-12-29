package secondary

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/review"
)

// IReviewRepository defines the interface for review data persistence
// All methods that access specific resources require userID to ensure data isolation
// Ownership is validated via JOIN with cards -> decks (reviews belong to cards, cards belong to decks, decks belong to users)
type IReviewRepository interface {
	// Save saves or updates a review in the database
	// If the review has an ID, it updates the existing review
	// If the review has no ID, it creates a new review and returns it with the ID set
	// Validates ownership via card_id -> deck_id
	Save(ctx context.Context, userID int64, reviewEntity *review.Review) error

	// FindByID finds a review by ID, filtering by userID via card ownership to ensure ownership
	// Returns the review if found and belongs to user's card, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*review.Review, error)

	// Update updates an existing review, validating ownership via card -> deck
	// Returns error if review doesn't exist or doesn't belong to user's card
	Update(ctx context.Context, userID int64, id int64, reviewEntity *review.Review) error

	// Delete deletes a review, validating ownership via card -> deck (hard delete - reviews don't have soft delete)
	// Returns error if review doesn't exist or doesn't belong to user's card
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a review exists and belongs to a user's card
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByCardID finds all reviews for a specific card, validating ownership
	FindByCardID(ctx context.Context, userID int64, cardID int64) ([]*review.Review, error)

	// FindByDateRange finds all reviews within a date range for a user
	FindByDateRange(ctx context.Context, userID int64, startDate time.Time, endDate time.Time) ([]*review.Review, error)
}

