package request

// CreateDeckOptionsPresetRequest represents the request payload to create a new deck options preset
// @Description Request payload for creating a new deck options preset
type CreateDeckOptionsPresetRequest struct {
	// Name of the preset
	Name string `json:"name" example:"Default Study" validate:"required"`

	// Configuration options for the preset (JSON string)
	OptionsJSON string `json:"options_json" example:"{}" validate:"required"`
}

// UpdateDeckOptionsPresetRequest represents the request payload to update an existing deck options preset
// @Description Request payload for updating an existing deck options preset
type UpdateDeckOptionsPresetRequest struct {
	// New name of the preset
	Name string `json:"name" example:"Hard Study" validate:"required"`

	// New configuration options for the preset (JSON string)
	OptionsJSON string `json:"options_json" example:"{}" validate:"required"`
}

// ApplyDeckOptionsPresetRequest represents the request payload to apply a preset to multiple decks
// @Description Request payload for applying a preset to multiple decks
type ApplyDeckOptionsPresetRequest struct {
	// IDs of the decks to apply the preset to
	DeckIDs []int64 `json:"deck_ids" example:"[1, 2, 3]" validate:"required,min=1"`
}

