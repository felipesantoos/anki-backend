package secondary

import (
	"context"

	addon "github.com/felipesantos/anki-backend/core/domain/entities/add_on"
)

// IAddOnRepository defines the interface for add-on data persistence
// All methods that access specific resources require userID to ensure data isolation
type IAddOnRepository interface {
	// Save saves or updates an add-on in the database
	// If the add-on has an ID, it updates the existing add-on
	// If the add-on has no ID, it creates a new add-on and returns it with the ID set
	Save(ctx context.Context, userID int64, addOnEntity *addon.AddOn) error

	// FindByID finds an add-on by ID, filtering by userID to ensure ownership
	// Returns the add-on if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*addon.AddOn, error)

	// FindByUserID finds all add-ons for a user
	// Returns a list of add-ons belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*addon.AddOn, error)

	// Update updates an existing add-on, validating ownership
	// Returns error if add-on doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, addOnEntity *addon.AddOn) error

	// Delete deletes an add-on, validating ownership (hard delete - add_ons don't have soft delete)
	// Returns error if add-on doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if an add-on exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByCode finds an add-on by code, filtering by userID to ensure ownership
	FindByCode(ctx context.Context, userID int64, code string) (*addon.AddOn, error)

	// FindEnabled finds all enabled add-ons for a user
	FindEnabled(ctx context.Context, userID int64) ([]*addon.AddOn, error)
}

