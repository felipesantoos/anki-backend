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

// ListNotesRequest represents the query parameters for listing notes
type ListNotesRequest struct {
	DeckID     *int64   `query:"deck_id"`
	NoteTypeID *int64   `query:"note_type_id"`
	Tags       []string `query:"tags"`
	Search     string   `query:"search"`
	Limit      int      `query:"limit"`
	Offset     int      `query:"offset"`
}

// CopyNoteRequest represents the request payload to copy a note
type CopyNoteRequest struct {
	// Optional deck ID (if not provided, uses original note's deck)
	DeckID *int64 `json:"deck_id" example:"1" validate:"omitempty,gt=0"`

	// Whether to copy tags from original note
	CopyTags bool `json:"copy_tags" example:"true"`

	// Whether to copy media from original note (future feature)
	CopyMedia bool `json:"copy_media" example:"true"`
}

// FindDuplicatesRequest represents the request payload to find duplicate notes
type FindDuplicatesRequest struct {
	// Optional note type ID (if not provided, searches across all note types)
	// If provided and field_name is empty, automatically uses the first field of the note type
	NoteTypeID *int64 `json:"note_type_id" example:"1" validate:"omitempty,gt=0"`

	// Field name to search for duplicates (optional)
	// If note_type_id is provided and field_name is empty, the first field of the note type will be used automatically
	// If note_type_id is not provided, field_name is required
	FieldName string `json:"field_name" example:"Front" validate:"omitempty"`

	// Use GUID for duplicate detection instead of field value
	// If true, finds duplicates based on GUID value (ignores field_name and note_type_id)
	UseGUID bool `json:"use_guid" example:"false"`
}

