package secondary

import (
	"context"

	flagname "github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
)

// IFlagNameRepository defines the interface for flag name data persistence
// All methods that access specific resources require userID to ensure data isolation
type IFlagNameRepository interface {
	// Save saves or updates a flag name in the database
	// If the flag name has an ID, it updates the existing flag name
	// If the flag name has no ID, it creates a new flag name and returns it with the ID set
	Save(ctx context.Context, userID int64, flagNameEntity *flagname.FlagName) error

	// FindByID finds a flag name by ID, filtering by userID to ensure ownership
	// Returns the flag name if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*flagname.FlagName, error)

	// FindByUserID finds all flag names for a user
	// Returns a list of flag names belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*flagname.FlagName, error)

	// Update updates an existing flag name, validating ownership
	// Returns error if flag name doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, flagNameEntity *flagname.FlagName) error

	// Delete deletes a flag name, validating ownership (hard delete - flag_names don't have soft delete)
	// Returns error if flag name doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a flag name exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)
}

