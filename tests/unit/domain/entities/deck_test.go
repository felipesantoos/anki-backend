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
			deck: &entities.Deck{
				DeletedAt: nil,
			},
			expected: true,
		},
		{
			name: "deleted deck",
			deck: &entities.Deck{
				DeletedAt: timePtr(time.Now()),
			},
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
			deck: &entities.Deck{
				ParentID: nil,
			},
			expected: true,
		},
		{
			name: "child deck",
			deck: &entities.Deck{
				ParentID: int64Ptr(1),
			},
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
	parent := &entities.Deck{
		ID:   1,
		Name: "Parent",
	}
	child := &entities.Deck{
		ID:       2,
		Name:     "Child",
		ParentID: int64Ptr(1),
	}
	grandchild := &entities.Deck{
		ID:       3,
		Name:     "Grandchild",
		ParentID: int64Ptr(2),
	}

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
			name:     "orphaned deck",
			deck:     &entities.Deck{ID: 4, Name: "Orphan", ParentID: int64Ptr(999)},
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
	deck := &entities.Deck{
		DeletedAt: nil,
	}

	if !deck.CanDelete() {
		t.Errorf("Deck.CanDelete() = false, want true for active deck")
	}

	deletedDeck := &entities.Deck{
		DeletedAt: timePtr(time.Now()),
	}

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
			deck: &entities.Deck{
				ParentID: int64Ptr(1),
			},
			expected: true,
		},
		{
			name: "no parent",
			deck: &entities.Deck{
				ParentID: nil,
			},
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


