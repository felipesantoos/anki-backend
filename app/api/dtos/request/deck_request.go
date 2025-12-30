package request

// CreateDeckRequest represents the request payload to create a new deck
// @Description Request payload for creating a new deck
type CreateDeckRequest struct {
	// Name of the deck
	Name string `json:"name" example:"Idiomas::Inglês" validate:"required"`

	// ID of the parent deck (optional)
	ParentID *int64 `json:"parent_id,omitempty" example:"1"`

	// Configuration options for the deck (JSON string)
	OptionsJSON string `json:"options_json,omitempty" example:"{}"`
}

// UpdateDeckRequest represents the request payload to update an existing deck
// @Description Request payload for updating an existing deck
type UpdateDeckRequest struct {
	// New name of the deck
	Name string `json:"name" example:"Idiomas::Inglês::Avançado" validate:"required"`

	// New parent deck ID (optional)
	ParentID *int64 `json:"parent_id,omitempty" example:"2"`

	// New configuration options for the deck (JSON string)
	OptionsJSON string `json:"options_json,omitempty" example:"{}"`
}

