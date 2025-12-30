package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
)

// IFlagNameService defines the interface for flag name management
type IFlagNameService interface {
	// FindByUserID finds all flag names for a user
	FindByUserID(ctx context.Context, userID int64) ([]*flagname.FlagName, error)

	// Update updates a flag name
	Update(ctx context.Context, userID int64, flagNumber int, name string) (*flagname.FlagName, error)
}

