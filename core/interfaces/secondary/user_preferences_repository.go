package secondary

import (
	"context"

	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
)

// IUserPreferencesRepository defines the interface for user preferences data persistence
// UserPreferences has a unique user_id constraint (one-to-one relationship with users)
type IUserPreferencesRepository interface {
	// Save saves or updates user preferences in the database
	// If the preferences have an ID, it updates the existing preferences
	// If the preferences have no ID, it creates new preferences and returns it with the ID set
	Save(ctx context.Context, userID int64, prefsEntity *userpreferences.UserPreferences) error

	// FindByID finds user preferences by ID, filtering by userID to ensure ownership
	// Returns the preferences if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*userpreferences.UserPreferences, error)

	// FindByUserID finds user preferences for a user (one-to-one relationship)
	// Returns the preferences if found, nil if not found, or an error
	FindByUserID(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error)

	// Update updates existing user preferences, validating ownership
	// Returns error if preferences don't exist or don't belong to user
	Update(ctx context.Context, userID int64, id int64, prefsEntity *userpreferences.UserPreferences) error

	// Delete deletes user preferences, validating ownership (hard delete - user_preferences doesn't have soft delete)
	// Returns error if preferences don't exist or don't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if user preferences exist for a user
	Exists(ctx context.Context, userID int64) (bool, error)
}

