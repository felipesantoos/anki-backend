package models

import (
	"database/sql"
	"time"
)

// DeckModel represents the decks table structure in the database
type DeckModel struct {
	ID          int64
	UserID      int64
	Name        string
	ParentID    sql.NullInt64
	OptionsJSON string // JSONB stored as string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

