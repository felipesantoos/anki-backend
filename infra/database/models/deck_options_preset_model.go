package models

import (
	"database/sql"
	"time"
)

// DeckOptionsPresetModel represents the deck_options_presets table structure in the database
type DeckOptionsPresetModel struct {
	ID          int64
	UserID      int64
	Name        string
	OptionsJSON string // JSONB stored as string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

