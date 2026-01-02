package response

import "time"

// NoteResponse represents the response payload for a note
type NoteResponse struct {
	// Unique identifier for the note
	ID int64 `json:"id" example:"1"`

	// GUID for synchronization
	GUID string `json:"guid" example:"abc-123"`

	// ID of the note type
	NoteTypeID int64 `json:"note_type_id" example:"1"`

	// Fields data as a JSON string
	FieldsJSON string `json:"fields_json" example:"{\"Front\": \"cat\", \"Back\": \"gato\"}"`

	// List of tags
	Tags []string `json:"tags" example:"[\"idiomas\"]"`

	// Timestamp when created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Timestamp when last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// FindDuplicatesResponse represents the response payload for finding duplicates
type FindDuplicatesResponse struct {
	// List of duplicate groups
	Duplicates []DuplicateGroup `json:"duplicates"`

	// Total number of duplicate groups found
	Total int `json:"total_duplicates" example:"1"`
}

// DuplicateGroup represents a group of duplicate notes with the same field value
type DuplicateGroup struct {
	// The field value that is duplicated (field value for field-based detection, GUID for GUID-based detection)
	FieldValue string `json:"field_value" example:"Hello"`

	// List of notes with this field value
	Notes []DuplicateNoteInfo `json:"notes"`
}

// DuplicateNoteInfo contains basic information about a duplicate note
type DuplicateNoteInfo struct {
	// Note ID
	ID int64 `json:"id" example:"1"`

	// Note GUID
	GUID string `json:"guid" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Deck ID where the note's cards are located
	DeckID int64 `json:"deck_id" example:"1"`

	// Timestamp when the note was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

