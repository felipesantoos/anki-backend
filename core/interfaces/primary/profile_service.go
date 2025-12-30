package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
)

// IProfileService defines the interface for profile management operations
type IProfileService interface {
	// Create creates a new profile for a user
	Create(ctx context.Context, userID int64, name string) (*profile.Profile, error)

	// FindByID finds a profile by ID
	FindByID(ctx context.Context, userID int64, id int64) (*profile.Profile, error)

	// FindByUserID finds all profiles for a user
	FindByUserID(ctx context.Context, userID int64) ([]*profile.Profile, error)

	// Update updates an existing profile
	Update(ctx context.Context, userID int64, id int64, name string) (*profile.Profile, error)

	// Delete deletes a profile (soft delete)
	Delete(ctx context.Context, userID int64, id int64) error

	// EnableSync enables AnkiWeb sync for a profile
	EnableSync(ctx context.Context, userID int64, id int64, username string) error

	// DisableSync disables AnkiWeb sync for a profile
	DisableSync(ctx context.Context, userID int64, id int64) error
}

