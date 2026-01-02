package notetype

import (
	"encoding/json"
	"fmt"
	"time"
)

// NoteType represents a note type entity in the domain
// It defines the structure and templates for notes
type NoteType struct {
	id            int64
	userID        int64
	name          string
	fieldsJSON    string // Array JSON in database
	cardTypesJSON string // Array JSON in database
	templatesJSON string // Object JSON in database
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
}

// Getters
func (nt *NoteType) GetID() int64 {
	return nt.id
}

func (nt *NoteType) GetUserID() int64 {
	return nt.userID
}

func (nt *NoteType) GetName() string {
	return nt.name
}

func (nt *NoteType) GetFieldsJSON() string {
	return nt.fieldsJSON
}

func (nt *NoteType) GetCardTypesJSON() string {
	return nt.cardTypesJSON
}

func (nt *NoteType) GetTemplatesJSON() string {
	return nt.templatesJSON
}

func (nt *NoteType) GetCreatedAt() time.Time {
	return nt.createdAt
}

func (nt *NoteType) GetUpdatedAt() time.Time {
	return nt.updatedAt
}

func (nt *NoteType) GetDeletedAt() *time.Time {
	return nt.deletedAt
}

// Setters
func (nt *NoteType) SetID(id int64) {
	nt.id = id
}

func (nt *NoteType) SetUserID(userID int64) {
	nt.userID = userID
}

func (nt *NoteType) SetName(name string) {
	nt.name = name
}

func (nt *NoteType) SetFieldsJSON(fieldsJSON string) {
	nt.fieldsJSON = fieldsJSON
}

func (nt *NoteType) SetCardTypesJSON(cardTypesJSON string) {
	nt.cardTypesJSON = cardTypesJSON
}

func (nt *NoteType) SetTemplatesJSON(templatesJSON string) {
	nt.templatesJSON = templatesJSON
}

func (nt *NoteType) SetCreatedAt(createdAt time.Time) {
	nt.createdAt = createdAt
}

func (nt *NoteType) SetUpdatedAt(updatedAt time.Time) {
	nt.updatedAt = updatedAt
}

func (nt *NoteType) SetDeletedAt(deletedAt *time.Time) {
	nt.deletedAt = deletedAt
}

// IsActive checks if the note type is active (not deleted)
func (nt *NoteType) IsActive() bool {
	return nt.deletedAt == nil
}

// GetFieldCount returns the number of fields in the note type
// Parses FieldsJSON to count fields
func (nt *NoteType) GetFieldCount() int {
	if nt.fieldsJSON == "" {
		return 0
	}

	var fields []interface{}
	if err := json.Unmarshal([]byte(nt.fieldsJSON), &fields); err != nil {
		return 0
	}

	return len(fields)
}

// GetCardTypeCount returns the number of card types in the note type
// Parses CardTypesJSON to count card types
func (nt *NoteType) GetCardTypeCount() int {
	if nt.cardTypesJSON == "" {
		return 0
	}

	var cardTypes []interface{}
	if err := json.Unmarshal([]byte(nt.cardTypesJSON), &cardTypes); err != nil {
		return 0
	}

	return len(cardTypes)
}

// GetFirstFieldName returns the name of the first field in the note type
// Returns error if fields array is empty or invalid
func (nt *NoteType) GetFirstFieldName() (string, error) {
	if nt.fieldsJSON == "" {
		return "", fmt.Errorf("note type has no fields defined")
	}

	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(nt.fieldsJSON), &fields); err != nil {
		return "", fmt.Errorf("invalid note type fields JSON: %w", err)
	}

	if len(fields) == 0 {
		return "", fmt.Errorf("note type has no fields defined")
	}

	// Get first field (index 0)
	firstField := fields[0]
	nameValue, exists := firstField["name"]
	if !exists {
		return "", fmt.Errorf("first field has no name property")
	}

	name, ok := nameValue.(string)
	if !ok {
		return "", fmt.Errorf("first field name is not a string")
	}

	if name == "" {
		return "", fmt.Errorf("first field name is empty")
	}

	return name, nil
}

