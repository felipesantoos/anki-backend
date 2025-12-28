package entities

import (
	"encoding/json"
	"time"
)

// NoteType represents a note type entity in the domain
// It defines the structure and templates for notes
type NoteType struct {
	ID            int64
	UserID        int64
	Name          string
	FieldsJSON    string // Array JSON in database
	CardTypesJSON string // Array JSON in database
	TemplatesJSON string // Object JSON in database
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// IsActive checks if the note type is active (not deleted)
func (nt *NoteType) IsActive() bool {
	return nt.DeletedAt == nil
}

// GetFieldCount returns the number of fields in the note type
// Parses FieldsJSON to count fields
func (nt *NoteType) GetFieldCount() int {
	if nt.FieldsJSON == "" {
		return 0
	}

	var fields []interface{}
	if err := json.Unmarshal([]byte(nt.FieldsJSON), &fields); err != nil {
		return 0
	}

	return len(fields)
}

// GetCardTypeCount returns the number of card types in the note type
// Parses CardTypesJSON to count card types
func (nt *NoteType) GetCardTypeCount() int {
	if nt.CardTypesJSON == "" {
		return 0
	}

	var cardTypes []interface{}
	if err := json.Unmarshal([]byte(nt.CardTypesJSON), &cardTypes); err != nil {
		return 0
	}

	return len(cardTypes)
}

