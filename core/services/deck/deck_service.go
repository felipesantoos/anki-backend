package deck

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// DeckService implements IDeckService
type DeckService struct {
	deckRepo secondary.IDeckRepository
}

// NewDeckService creates a new DeckService instance
func NewDeckService(deckRepo secondary.IDeckRepository) primary.IDeckService {
	return &DeckService{
		deckRepo: deckRepo,
	}
}

// Create creates a new deck for a user
func (s *DeckService) Create(ctx context.Context, userID int64, name string, parentID *int64, optionsJSON string) (*deck.Deck, error) {
	// 1. Check if deck with same name exists at same level
	exists, err := s.deckRepo.Exists(ctx, userID, name, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check deck existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("deck with name %s already exists at this level", name)
	}

	// 2. Validate parent if provided
	if parentID != nil {
		parent, err := s.deckRepo.FindByID(ctx, userID, *parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent deck: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent deck not found")
		}
	}

	// 3. Create deck entity using builder
	now := time.Now()
	if optionsJSON == "" {
		optionsJSON = "{}"
	}

	deckEntity, err := deck.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName(name).
		WithParentID(parentID).
		WithOptionsJSON(optionsJSON).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build deck entity: %w", err)
	}

	// 4. Save to repository
	if err := s.deckRepo.Save(ctx, userID, deckEntity); err != nil {
		return nil, fmt.Errorf("failed to save deck: %w", err)
	}

	return deckEntity, nil
}

// FindByID finds a deck by ID, validating ownership
func (s *DeckService) FindByID(ctx context.Context, userID int64, id int64) (*deck.Deck, error) {
	return s.deckRepo.FindByID(ctx, userID, id)
}

// FindByUserID finds all decks for a user
func (s *DeckService) FindByUserID(ctx context.Context, userID int64) ([]*deck.Deck, error) {
	return s.deckRepo.FindByUserID(ctx, userID)
}

// Update updates an existing deck
func (s *DeckService) Update(ctx context.Context, userID int64, id int64, name string, parentID *int64, optionsJSON string) (*deck.Deck, error) {
	// 1. Find existing deck
	existing, err := s.deckRepo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("deck not found")
	}

	// 2. If name or parent changed, check for conflicts
	nameChanged := existing.GetName() != name
	parentChanged := false
	if existing.GetParentID() != nil && parentID == nil {
		parentChanged = true
	} else if existing.GetParentID() == nil && parentID != nil {
		parentChanged = true
	} else if existing.GetParentID() != nil && parentID != nil && *existing.GetParentID() != *parentID {
		parentChanged = true
	}

	if nameChanged || parentChanged {
		exists, err := s.deckRepo.Exists(ctx, userID, name, parentID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("deck with name %s already exists at this level", name)
		}
	}

	// 3. Validate parent if changed
	if parentChanged && parentID != nil {
		// Prevent circular dependency (simplified: just check if parent is not self)
		if *parentID == id {
			return nil, fmt.Errorf("deck cannot be its own parent")
		}
		parent, err := s.deckRepo.FindByID(ctx, userID, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, fmt.Errorf("parent deck not found")
		}
	}

	// 4. Update entity
	existing.SetName(name)
	existing.SetParentID(parentID)
	if optionsJSON != "" {
		existing.SetOptionsJSON(optionsJSON)
	}
	existing.SetUpdatedAt(time.Now())

	// 5. Save
	if err := s.deckRepo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a deck (soft delete)
func (s *DeckService) Delete(ctx context.Context, userID int64, id int64) error {
	// 1. Find deck
	existing, err := s.deckRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("deck not found")
	}

	// 2. Prevent deleting default deck
	if existing.GetName() == "Default" && existing.GetParentID() == nil {
		return fmt.Errorf("cannot delete the default deck")
	}

	// 3. Perform soft delete
	return s.deckRepo.Delete(ctx, userID, id)
}

// CreateDefaultDeck creates the initial "Default" deck for a user
func (s *DeckService) CreateDefaultDeck(ctx context.Context, userID int64) (*deck.Deck, error) {
	deckID, err := s.deckRepo.CreateDefaultDeck(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.deckRepo.FindByID(ctx, userID, deckID)
}

