package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
)

// IProfileRepository defines the interface for profile data persistence
// All methods that access specific resources require userID to ensure data isolation
type IProfileRepository interface {
	// Save saves or updates a profile in the database
	// If the profile has an ID, it updates the existing profile
	// If the profile has no ID, it creates a new profile and returns it with the ID set
	Save(ctx context.Context, userID int64, profileEntity *profile.Profile) error

	// FindByID finds a profile by ID, filtering by userID to ensure ownership
	// Returns the profile if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*profile.Profile, error)

	// FindByUserID finds all profiles for a user
	// Returns a list of profiles belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*profile.Profile, error)

	// Update updates an existing profile, validating ownership
	// Returns error if profile doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, profileEntity *profile.Profile) error

	// Delete deletes a profile, validating ownership (soft delete)
	// Returns error if profile doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a profile exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)
}

