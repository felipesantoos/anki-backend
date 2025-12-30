package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
)

// IUserPreferencesService defines the interface for user preferences management
type IUserPreferencesService interface {
	// FindByUserID finds preferences for a user
	FindByUserID(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error)

	// Update updates user preferences
	Update(ctx context.Context, userID int64, prefs *userpreferences.UserPreferences) error

	// ResetToDefaults resets user preferences to default values
	ResetToDefaults(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error)
}

