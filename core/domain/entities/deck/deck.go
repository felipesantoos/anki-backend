package deck

import (
	"time"
)

// Deck represents a deck (card deck) entity in the domain
// It contains the core business logic for deck management
type Deck struct {
	id          int64
	userID      int64
	name        string
	parentID    *int64
	optionsJSON string // JSONB in database
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// DeleteAction defines the strategy for handling cards when deleting a deck
type DeleteAction string

const (
	// ActionDeleteCards permanently deletes all cards in the deck
	ActionDeleteCards DeleteAction = "delete_cards"
	// ActionMoveToDefault moves all cards to the user's default deck
	ActionMoveToDefault DeleteAction = "move_to_default"
	// ActionMoveToDeck moves all cards to a specific target deck
	ActionMoveToDeck DeleteAction = "move_to_deck"
)

// Getters
func (d *Deck) GetID() int64 {
	return d.id
}

func (d *Deck) GetUserID() int64 {
	return d.userID
}

func (d *Deck) GetName() string {
	return d.name
}

func (d *Deck) GetParentID() *int64 {
	return d.parentID
}

func (d *Deck) GetOptionsJSON() string {
	return d.optionsJSON
}

func (d *Deck) GetCreatedAt() time.Time {
	return d.createdAt
}

func (d *Deck) GetUpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Deck) GetDeletedAt() *time.Time {
	return d.deletedAt
}

// Setters
func (d *Deck) SetID(id int64) {
	d.id = id
}

func (d *Deck) SetUserID(userID int64) {
	d.userID = userID
}

func (d *Deck) SetName(name string) {
	d.name = name
}

func (d *Deck) SetParentID(parentID *int64) {
	d.parentID = parentID
}

func (d *Deck) SetOptionsJSON(optionsJSON string) {
	d.optionsJSON = optionsJSON
}

func (d *Deck) SetCreatedAt(createdAt time.Time) {
	d.createdAt = createdAt
}

func (d *Deck) SetUpdatedAt(updatedAt time.Time) {
	d.updatedAt = updatedAt
}

func (d *Deck) SetDeletedAt(deletedAt *time.Time) {
	d.deletedAt = deletedAt
}

// IsActive checks if the deck is active (not deleted)
func (d *Deck) IsActive() bool {
	return d.deletedAt == nil
}

// IsRoot checks if the deck is a root deck (has no parent)
func (d *Deck) IsRoot() bool {
	return d.parentID == nil
}

// GetFullPath returns the full hierarchical path of the deck
// Example: "Parent::Child::Grandchild"
// decks parameter should contain all decks in the hierarchy
func (d *Deck) GetFullPath(decks []*Deck) string {
	path := []string{d.GetName()}
	current := d

	// Build path by traversing up the hierarchy
	for current.GetParentID() != nil {
		found := false
		for _, deck := range decks {
			if deck.GetID() == *current.GetParentID() {
				path = append([]string{deck.GetName()}, path...)
				current = deck
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// Join with "::" separator
	result := ""
	for i, name := range path {
		if i > 0 {
			result += "::"
		}
		result += name
	}
	return result
}

// CanDelete checks if the deck can be deleted
// This is a domain method - actual validation should check if deck has cards
// Returns true if deck is active (actual card count check should be done in service layer)
func (d *Deck) CanDelete() bool {
	return d.IsActive()
}

// HasParent checks if the deck has a parent
func (d *Deck) HasParent() bool {
	return d.parentID != nil
}

