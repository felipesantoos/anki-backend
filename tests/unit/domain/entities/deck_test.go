package entities

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
)

func TestDeck_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		deck     *entities.Deck
		expected bool
	}{
		{
			name: "active deck",
			deck: func() *entities.Deck {
				d := &entities.Deck{}
				d.SetDeletedAt(nil)
				return d
			}(),
			expected: true,
		},
		{
			name: "deleted deck",
			deck: func() *entities.Deck {
				d := &entities.Deck{}
				d.SetDeletedAt(timePtr(time.Now()))
				return d
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.deck.IsActive()
			if got != tt.expected {
				t.Errorf("Deck.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDeck_IsRoot(t *testing.T) {
	tests := []struct {
		name     string
		deck     *entities.Deck
		expected bool
	}{
		{
			name: "root deck",
			deck: func() *entities.Deck {
				d := &entities.Deck{}
				d.SetParentID(nil)
				return d
			}(),
			expected: true,
		},
		{
			name: "child deck",
			deck: func() *entities.Deck {
				d := &entities.Deck{}
				d.SetParentID(int64Ptr(1))
				return d
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.deck.IsRoot()
			if got != tt.expected {
				t.Errorf("Deck.IsRoot() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDeck_GetFullPath(t *testing.T) {
	parent := &entities.Deck{}
	parent.SetID(1)
	parent.SetName("Parent")
	
	child := &entities.Deck{}
	child.SetID(2)
	child.SetName("Child")
	child.SetParentID(int64Ptr(1))
	
	grandchild := &entities.Deck{}
	grandchild.SetID(3)
	grandchild.SetName("Grandchild")
	grandchild.SetParentID(int64Ptr(2))

	allDecks := []*entities.Deck{parent, child, grandchild}

	tests := []struct {
		name     string
		deck     *entities.Deck
		decks    []*entities.Deck
		expected string
	}{
		{
			name:     "root deck",
			deck:     parent,
			decks:    allDecks,
			expected: "Parent",
		},
		{
			name:     "child deck",
			deck:     child,
			decks:    allDecks,
			expected: "Parent::Child",
		},
		{
			name:     "grandchild deck",
			deck:     grandchild,
			decks:    allDecks,
			expected: "Parent::Child::Grandchild",
		},
		{
			name: "orphaned deck",
			deck: func() *entities.Deck {
				d := &entities.Deck{}
				d.SetID(4)
				d.SetName("Orphan")
				d.SetParentID(int64Ptr(999))
				return d
			}(),
			decks:    allDecks,
			expected: "Orphan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.deck.GetFullPath(tt.decks)
			if got != tt.expected {
				t.Errorf("Deck.GetFullPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDeck_CanDelete(t *testing.T) {
	deck := &entities.Deck{}
	deck.SetDeletedAt(nil)

	if !deck.CanDelete() {
		t.Errorf("Deck.CanDelete() = false, want true for active deck")
	}

	deletedDeck := &entities.Deck{}
	deletedDeck.SetDeletedAt(timePtr(time.Now()))

	if deletedDeck.CanDelete() {
		t.Errorf("Deck.CanDelete() = true, want false for deleted deck")
	}
}

func TestDeck_HasParent(t *testing.T) {
	tests := []struct {
		name     string
		deck     *entities.Deck
		expected bool
	}{
		{
			name: "has parent",
			deck: func() *entities.Deck {
				d := &entities.Deck{}
				d.SetParentID(int64Ptr(1))
				return d
			}(),
			expected: true,
		},
		{
			name: "no parent",
			deck: func() *entities.Deck {
				d := &entities.Deck{}
				d.SetParentID(nil)
				return d
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.deck.HasParent()
			if got != tt.expected {
				t.Errorf("Deck.HasParent() = %v, want %v", got, tt.expected)
			}
		})
	}
}


