package request

// CreateNoteTypeRequest represents the request payload to create a new note type
type CreateNoteTypeRequest struct {
	// Name of the note type
	Name string `json:"name" example:"Básico" validate:"required"`

	// Field definitions as a JSON string
	FieldsJSON string `json:"fields_json" example:"[{\"name\": \"Front\"}, {\"name\": \"Back\"}]" validate:"required"`

	// Card type definitions as a JSON string
	CardTypesJSON string `json:"card_types_json" example:"[{\"name\": \"Cartão 1\"}]" validate:"required"`

	// Templates definitions as a JSON string
	TemplatesJSON string `json:"templates_json" example:"{\"qfmt\": \"{{Front}}\", \"afmt\": \"{{FrontSide}}<hr id=answer>{{Back}}\"}" validate:"required"`
}

// UpdateNoteTypeRequest represents the request payload to update an existing note type
type UpdateNoteTypeRequest struct {
	// New name of the note type
	Name string `json:"name" example:"Básico com Reverso" validate:"required"`

	// Updated field definitions
	FieldsJSON string `json:"fields_json" validate:"required"`

	// Updated card type definitions
	CardTypesJSON string `json:"card_types_json" validate:"required"`

	// Updated templates definitions
	TemplatesJSON string `json:"templates_json" validate:"required"`
}

// ListNoteTypesRequest represents the query parameters for listing note types
type ListNoteTypesRequest struct {
	// Search term to filter note types by name (case-insensitive partial matching)
	Search string `query:"search" example:"Basic"`
}

