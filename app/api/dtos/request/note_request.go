package request

// CreateNoteRequest represents the request payload to create a new note
type CreateNoteRequest struct {
	// ID of the note type (template)
	NoteTypeID int64 `json:"note_type_id" example:"1" validate:"required"`

	// ID of the deck where cards will be created
	DeckID int64 `json:"deck_id" example:"1" validate:"required"`

	// Fields data as a JSON string (e.g., '{"Front": "Hello", "Back": "Olá"}')
	FieldsJSON string `json:"fields_json" example:"{\"Front\": \"cat\", \"Back\": \"gato\"}" validate:"required"`

	// List of tags for the note
	Tags []string `json:"tags" example:"[\"idiomas\", \"animal\"]"`
}

// UpdateNoteRequest represents the request payload to update an existing note
type UpdateNoteRequest struct {
	// Updated fields data as a JSON string
	FieldsJSON string `json:"fields_json" example:"{\"Front\": \"cat\", \"Back\": \"gato (felino)\"}" validate:"required"`

	// Updated list of tags
	Tags []string `json:"tags" example:"[\"idiomas\", \"mamífero\"]"`
}

// AddTagRequest represents the request payload to add a tag
type AddTagRequest struct {
	// Tag name to add
	Tag string `json:"tag" example:"importante" validate:"required"`
}

