package models

import (
	"database/sql"
	"time"
)

// NoteTypeModel represents the note_types table structure in the database
type NoteTypeModel struct {
	ID            int64
	UserID        int64
	Name          string
	FieldsJSON    string // JSONB stored as string
	CardTypesJSON string // JSONB stored as string
	TemplatesJSON string // JSONB stored as string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     sql.NullTime
}

