package models

import (
	"database/sql"
	"time"
)

// SharedDeckRatingModel represents the shared_deck_ratings table structure in the database
type SharedDeckRatingModel struct {
	ID           int64
	UserID       int64
	SharedDeckID int64
	Rating       int
	Comment      sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

