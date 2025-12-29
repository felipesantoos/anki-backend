package secondary

import (
	"context"

	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
)

// ISharedDeckRatingRepository defines the interface for shared deck rating data persistence
// All methods that access specific resources require userID to ensure data isolation
type ISharedDeckRatingRepository interface {
	// Save saves or updates a shared deck rating in the database
	// If the rating has an ID, it updates the existing rating
	// If the rating has no ID, it creates a new rating and returns it with the ID set
	Save(ctx context.Context, userID int64, ratingEntity *shareddeckrating.SharedDeckRating) error

	// FindByID finds a shared deck rating by ID, filtering by userID to ensure ownership
	// Returns the rating if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*shareddeckrating.SharedDeckRating, error)

	// FindByUserID finds all shared deck ratings by a user
	// Returns a list of ratings belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*shareddeckrating.SharedDeckRating, error)

	// Update updates an existing shared deck rating, validating ownership
	// Returns error if rating doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, ratingEntity *shareddeckrating.SharedDeckRating) error

	// Delete deletes a shared deck rating, validating ownership (hard delete - shared_deck_ratings doesn't have soft delete)
	// Returns error if rating doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a shared deck rating exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// Specific methods
	// FindBySharedDeckID finds all ratings for a shared deck
	FindBySharedDeckID(ctx context.Context, sharedDeckID int64) ([]*shareddeckrating.SharedDeckRating, error)

	// FindByUserIDAndSharedDeckID finds a rating by user and shared deck (one rating per user per deck)
	FindByUserIDAndSharedDeckID(ctx context.Context, userID int64, sharedDeckID int64) (*shareddeckrating.SharedDeckRating, error)
}

