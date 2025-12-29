package models

import (
	"database/sql"
	"time"
)

// ProfileModel represents the profiles table structure in the database
type ProfileModel struct {
	ID                 int64
	UserID             int64
	Name               string
	AnkiWebSyncEnabled bool
	AnkiWebUsername    sql.NullString
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt         sql.NullTime
}

