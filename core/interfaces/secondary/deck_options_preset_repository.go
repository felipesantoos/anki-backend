package secondary

import (
	"context"

	deckoptionspreset "github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
)

// IDeckOptionsPresetRepository defines the interface for deck options preset data persistence
// All methods that access specific resources require userID to ensure data isolation
type IDeckOptionsPresetRepository interface {
	// Save saves or updates a deck options preset in the database
	// If the preset has an ID, it updates the existing preset
	// If the preset has no ID, it creates a new preset and returns it with the ID set
	Save(ctx context.Context, userID int64, presetEntity *deckoptionspreset.DeckOptionsPreset) error

	// FindByID finds a deck options preset by ID, filtering by userID to ensure ownership
	// Returns the preset if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*deckoptionspreset.DeckOptionsPreset, error)

	// FindByUserID finds all deck options presets for a user
	// Returns a list of presets belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*deckoptionspreset.DeckOptionsPreset, error)

	// Update updates an existing deck options preset, validating ownership
	// Returns error if preset doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, presetEntity *deckoptionspreset.DeckOptionsPreset) error

	// Delete deletes a deck options preset, validating ownership (soft delete)
	// Returns error if preset doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a deck options preset exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByName finds a deck options preset by name, filtering by userID to ensure ownership
	FindByName(ctx context.Context, userID int64, name string) (*deckoptionspreset.DeckOptionsPreset, error)
}

