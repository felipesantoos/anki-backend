package models

import (
	"database/sql"
	"time"
)

// CardModel represents the cards table structure in the database
// It uses database/sql nullable types for optional fields
type CardModel struct {
	ID           int64
	NoteID       int64
	CardTypeID   int
	DeckID       int64
	HomeDeckID   sql.NullInt64
	Due          int64
	Interval     int
	Ease         int
	Lapses       int
	Reps         int
	State        string // card_state enum stored as string
	Position     int
	Flag         int
	Suspended    bool
	Buried       bool
	Stability    sql.NullFloat64
	Difficulty   sql.NullFloat64
	LastReviewAt sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

