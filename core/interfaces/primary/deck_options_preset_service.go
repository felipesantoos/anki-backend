package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
)

// IDeckOptionsPresetService defines the interface for deck options preset management
type IDeckOptionsPresetService interface {
	// Create creates a new options preset
	Create(ctx context.Context, userID int64, name string, optionsJSON string) (*deckoptionspreset.DeckOptionsPreset, error)

	// FindByUserID finds all presets for a user
	FindByUserID(ctx context.Context, userID int64) ([]*deckoptionspreset.DeckOptionsPreset, error)

	// Update updates an existing preset
	Update(ctx context.Context, userID int64, id int64, name string, optionsJSON string) (*deckoptionspreset.DeckOptionsPreset, error)

	// Delete deletes a preset
	Delete(ctx context.Context, userID int64, id int64) error

	// ApplyToDecks applies a preset's options to multiple decks
	ApplyToDecks(ctx context.Context, userID int64, id int64, deckIDs []int64) error
}

