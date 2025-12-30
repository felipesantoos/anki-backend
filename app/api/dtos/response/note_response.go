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

