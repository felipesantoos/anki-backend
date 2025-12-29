package models

import (
	"database/sql"
	"time"
)

// FilteredDeckModel represents the filtered_decks table structure in the database
type FilteredDeckModel struct {
	ID            int64
	UserID        int64
	Name          string
	SearchFilter  string
	SecondFilter  sql.NullString
	LimitCards    int
	OrderBy       string
	Reschedule    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastRebuildAt sql.NullTime
	DeletedAt     sql.NullTime
}

