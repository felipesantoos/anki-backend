package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
)

// IBrowserConfigService defines the interface for browser config management
type IBrowserConfigService interface {
	// FindByUserID finds browser config for a user
	FindByUserID(ctx context.Context, userID int64) (*browserconfig.BrowserConfig, error)

	// Update updates browser config
	Update(ctx context.Context, userID int64, config *browserconfig.BrowserConfig) error
}

