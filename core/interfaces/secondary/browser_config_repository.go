package secondary

import (
	"context"

	browserconfig "github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
)

// IBrowserConfigRepository defines the interface for browser config data persistence
// BrowserConfig has a unique user_id constraint (one-to-one relationship with users)
type IBrowserConfigRepository interface {
	// Save saves or updates browser config in the database
	// If the config has an ID, it updates the existing config
	// If the config has no ID, it creates new config and returns it with the ID set
	Save(ctx context.Context, userID int64, browserConfigEntity *browserconfig.BrowserConfig) error

	// FindByID finds browser config by ID, filtering by userID to ensure ownership
	// Returns the config if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*browserconfig.BrowserConfig, error)

	// FindByUserID finds browser config for a user (one-to-one relationship)
	// Returns the config if found, nil if not found, or an error
	FindByUserID(ctx context.Context, userID int64) (*browserconfig.BrowserConfig, error)

	// Update updates existing browser config, validating ownership
	// Returns error if config doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, browserConfigEntity *browserconfig.BrowserConfig) error

	// Delete deletes browser config, validating ownership (hard delete - browser_config doesn't have soft delete)
	// Returns error if config doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if browser config exists for a user
	Exists(ctx context.Context, userID int64) (bool, error)
}

