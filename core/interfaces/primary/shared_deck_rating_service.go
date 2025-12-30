package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
)

// ISharedDeckRatingService defines the interface for shared deck ratings
type ISharedDeckRatingService interface {
	// Create records a new rating for a shared deck
	Create(ctx context.Context, userID int64, sharedDeckID int64, rating int, comment *string) (*shareddeckrating.SharedDeckRating, error)

	// FindBySharedDeckID finds all ratings for a shared deck
	FindBySharedDeckID(ctx context.Context, sharedDeckID int64) ([]*shareddeckrating.SharedDeckRating, error)

	// Update updates an existing rating
	Update(ctx context.Context, userID int64, sharedDeckID int64, rating int, comment *string) (*shareddeckrating.SharedDeckRating, error)

	// Delete removes a rating
	Delete(ctx context.Context, userID int64, sharedDeckID int64) error
}

