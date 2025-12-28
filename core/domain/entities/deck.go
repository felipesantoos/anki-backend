package entities

import (
	"time"
)

// Deck represents a deck (card deck) entity in the domain
// It contains the core business logic for deck management
type Deck struct {
	ID          int64
	UserID      int64
	Name        string
	ParentID    *int64
	OptionsJSON string // JSONB in database
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// IsActive checks if the deck is active (not deleted)
func (d *Deck) IsActive() bool {
	return d.DeletedAt == nil
}

// IsRoot checks if the deck is a root deck (has no parent)
func (d *Deck) IsRoot() bool {
	return d.ParentID == nil
}

// GetFullPath returns the full hierarchical path of the deck
// Example: "Parent::Child::Grandchild"
// decks parameter should contain all decks in the hierarchy
func (d *Deck) GetFullPath(decks []*Deck) string {
	path := []string{d.Name}
	current := d

	// Build path by traversing up the hierarchy
	for current.ParentID != nil {
		found := false
		for _, deck := range decks {
			if deck.ID == *current.ParentID {
				path = append([]string{deck.Name}, path...)
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
	return d.ParentID != nil
}

