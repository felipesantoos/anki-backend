package entities

import (
	"time"
)

// DeckOptionsPreset represents a deck options preset entity in the domain
// It stores reusable deck configuration presets
type DeckOptionsPreset struct {
	ID          int64
	UserID      int64
	Name        string
	OptionsJSON string // JSONB in database
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// IsActive checks if the preset is active (not deleted)
func (dop *DeckOptionsPreset) IsActive() bool {
	return dop.DeletedAt == nil
}

