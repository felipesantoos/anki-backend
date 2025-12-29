package entities

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
)

func TestDeck_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		deck     *deck.Deck
		expected bool
	}{
		{
			name: "active deck",
			deck: func() *deck.Deck {
				d, _ := deck.NewBuilder().WithUserID(1).WithName("Test Deck").WithDeletedAt(nil).Build()
				return d
			}(),
			expected: true,
		},
		{
			name: "deleted deck",
			deck: func() *deck.Deck {
				now := time.Now()
				d, _ := deck.NewBuilder().WithUserID(1).WithName("Test Deck").WithDeletedAt(&now).Build()
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
		deck     *deck.Deck
		expected bool
	}{
		{
			name: "root deck",
			deck: func() *deck.Deck {
				d, _ := deck.NewBuilder().WithUserID(1).WithName("Root Deck").WithParentID(nil).Build()
				return d
			}(),
			expected: true,
		},
		{
			name: "child deck",
			deck: func() *deck.Deck {
				parentID := int64(1)
				d, _ := deck.NewBuilder().WithUserID(1).WithName("Child Deck").WithParentID(&parentID).Build()
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
	parent, _ := deck.NewBuilder().WithID(1).WithUserID(1).WithName("Parent").Build()
	
	parentID := int64(1)
	child, _ := deck.NewBuilder().WithID(2).WithUserID(1).WithName("Child").WithParentID(&parentID).Build()
	
	childID := int64(2)
	grandchild, _ := deck.NewBuilder().WithID(3).WithUserID(1).WithName("Grandchild").WithParentID(&childID).Build()

	allDecks := []*deck.Deck{parent, child, grandchild}

	tests := []struct {
		name     string
		deck     *deck.Deck
		decks    []*deck.Deck
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
			deck: func() *deck.Deck {
				orphanParentID := int64(999)
				d, _ := deck.NewBuilder().WithID(4).WithUserID(1).WithName("Orphan").WithParentID(&orphanParentID).Build()
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
	d, _ := deck.NewBuilder().WithUserID(1).WithName("Test Deck").WithDeletedAt(nil).Build()

	if !d.CanDelete() {
		t.Errorf("Deck.CanDelete() = false, want true for active deck")
	}

	now := time.Now()
	deletedDeck, _ := deck.NewBuilder().WithUserID(1).WithName("Test Deck").WithDeletedAt(&now).Build()

	if deletedDeck.CanDelete() {
		t.Errorf("Deck.CanDelete() = true, want false for deleted deck")
	}
}

func TestDeck_HasParent(t *testing.T) {
	tests := []struct {
		name     string
		deck     *deck.Deck
		expected bool
	}{
		{
			name: "has parent",
			deck: func() *deck.Deck {
				parentID := int64(1)
				d, _ := deck.NewBuilder().WithUserID(1).WithName("Child Deck").WithParentID(&parentID).Build()
				return d
			}(),
			expected: true,
		},
		{
			name: "no parent",
			deck: func() *deck.Deck {
				d, _ := deck.NewBuilder().WithUserID(1).WithName("Root Deck").WithParentID(nil).Build()
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


