package models

import (
	"database/sql"
	"time"
)

// SharedDeckModel represents the shared_decks table structure in the database
type SharedDeckModel struct {
	ID            int64
	AuthorID      int64
	Name          string
	Description   sql.NullString
	Category      sql.NullString
	PackagePath   string
	PackageSize   int64
	DownloadCount int
	RatingAverage float64
	RatingCount   int
	Tags          sql.NullString // TEXT[] stored as string
	IsFeatured    bool
	IsPublic      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime
}

