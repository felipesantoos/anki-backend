package models

import (
	"database/sql"
	"time"
)

// NoteModel represents the notes table structure in the database
// It uses database/sql nullable types for optional fields
type NoteModel struct {
	ID         int64
	UserID     int64
	GUID       string
	NoteTypeID int64
	FieldsJSON string // JSONB stored as string
	Tags       sql.NullString // TEXT[] array stored as string
	Marked     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  sql.NullTime
}

