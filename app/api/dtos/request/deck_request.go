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

// DeleteDeckAction defines the strategy for handling cards when deleting a deck
type DeleteDeckAction string

const (
	// ActionDeleteCards permanently deletes all cards in the deck
	ActionDeleteCards DeleteDeckAction = "delete_cards"
	// ActionMoveToDefault moves all cards to the user's default deck
	ActionMoveToDefault DeleteDeckAction = "move_to_default"
	// ActionMoveToDeck moves all cards to a specific target deck
	ActionMoveToDeck DeleteDeckAction = "move_to_deck"
)

// DeleteDeckRequest represents the request payload to delete a deck with card management choices
// @Description Request payload for deleting a deck with card management choices
type DeleteDeckRequest struct {
	// Strategy for handling cards in the deleted deck
	Action DeleteDeckAction `json:"action" example:"move_to_default" validate:"required,oneof=delete_cards move_to_default move_to_deck"`

	// ID of the target deck when action is 'move_to_deck'
	TargetDeckID *int64 `json:"target_deck_id,omitempty" example:"10"`
}

// ListDecksRequest represents the query parameters for listing decks
// @Description Query parameters for listing decks with optional search
type ListDecksRequest struct {
	// Search term to filter decks by name (case-insensitive partial matching)
	Search string `query:"search" example:"Math"`
}

