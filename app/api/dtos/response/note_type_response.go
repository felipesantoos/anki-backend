package response

import "time"

// NoteTypeResponse represents the response payload for a note type
type NoteTypeResponse struct {
	// Unique identifier for the note type
	ID int64 `json:"id" example:"1"`

	// ID of the user who owns the note type
	UserID int64 `json:"user_id" example:"1"`

	// Name of the note type
	Name string `json:"name" example:"Básico"`

	// Field definitions
	FieldsJSON string `json:"fields_json"`

	// Card type definitions
	CardTypesJSON string `json:"card_types_json"`

	// Templates definitions
	TemplatesJSON string `json:"templates_json"`

	// Timestamp when created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Timestamp when last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

