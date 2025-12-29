package models

import (
	"database/sql"
	"time"
)

// MediaModel represents the media table structure in the database
type MediaModel struct {
	ID          int64
	UserID      int64
	Filename    string
	Hash        string
	Size        int64
	MimeType    string
	StoragePath string
	CreatedAt   time.Time
	DeletedAt   sql.NullTime
}

