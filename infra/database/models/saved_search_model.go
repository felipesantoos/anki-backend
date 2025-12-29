package models

import (
	"database/sql"
	"time"
)

// SavedSearchModel represents the saved_searches table structure in the database
type SavedSearchModel struct {
	ID         int64
	UserID     int64
	Name       string
	SearchQuery string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  sql.NullTime
}

