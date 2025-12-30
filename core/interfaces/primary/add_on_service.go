package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/add_on"
)

// IAddOnService defines the interface for add-on management
type IAddOnService interface {
	// Install records a new add-on installation
	Install(ctx context.Context, userID int64, code string, name string, version string, configJSON string) (*addon.AddOn, error)

	// FindByUserID finds all add-ons for a user
	FindByUserID(ctx context.Context, userID int64) ([]*addon.AddOn, error)

	// UpdateConfig updates an add-on's configuration
	UpdateConfig(ctx context.Context, userID int64, code string, configJSON string) (*addon.AddOn, error)

	// ToggleEnabled enables or disables an add-on
	ToggleEnabled(ctx context.Context, userID int64, code string, enabled bool) (*addon.AddOn, error)

	// Uninstall removes an add-on
	Uninstall(ctx context.Context, userID int64, code string) error
}

