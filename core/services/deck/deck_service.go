package deck

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

var (
	// ErrDeckNotFound is returned when a deck cannot be found
	ErrDeckNotFound = errors.New("deck not found")
	// ErrCircularDependency is returned when a deck is moved into its own descendant
	ErrCircularDependency = errors.New("cannot move deck into itself or its own descendant")
)

// DeckService implements IDeckService
type DeckService struct {
	deckRepo   secondary.IDeckRepository
	cardRepo   secondary.ICardRepository
	backupSvc  primary.IBackupService
	tm         secondary.ITransactionManager
}

// NewDeckService creates a new DeckService instance
func NewDeckService(
	deckRepo secondary.IDeckRepository,
	cardRepo secondary.ICardRepository,
	backupSvc primary.IBackupService,
	tm secondary.ITransactionManager,
) primary.IDeckService {
	return &DeckService{
		deckRepo:   deckRepo,
		cardRepo:   cardRepo,
		backupSvc:  backupSvc,
		tm:         tm,
	}
}

// Create creates a new deck for a user
func (s *DeckService) Create(ctx context.Context, userID int64, name string, parentID *int64, optionsJSON string) (*deck.Deck, error) {
	// 0. Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("deck name cannot be empty")
	}
	if strings.Contains(name, "::") {
		return nil, fmt.Errorf("deck name cannot contain '::'")
	}

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
	// 0. Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("deck name cannot be empty")
	}
	if strings.Contains(name, "::") {
		return nil, fmt.Errorf("deck name cannot contain '::'")
	}

	// 1. Find existing deck
	existing, err := s.deckRepo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrDeckNotFound
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
		// Deep cycle check: ensure new parent is not the deck itself or one of its descendants
		tempParentID := parentID
		for tempParentID != nil {
			if *tempParentID == id {
				return nil, ErrCircularDependency
			}
			p, err := s.deckRepo.FindByID(ctx, userID, *tempParentID)
			if err != nil {
				return nil, err
			}
			if p == nil {
				return nil, fmt.Errorf("parent deck not found")
			}
			tempParentID = p.GetParentID()
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

// UpdateOptions updates only the options of an existing deck
func (s *DeckService) UpdateOptions(ctx context.Context, userID int64, id int64, optionsJSON string) (*deck.Deck, error) {
	// 1. Find existing deck (validates ownership)
	existing, err := s.deckRepo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrDeckNotFound
	}

	// 2. Update options
	if optionsJSON == "" {
		optionsJSON = "{}"
	}
	existing.SetOptionsJSON(optionsJSON)
	existing.SetUpdatedAt(time.Now())

	// 3. Save
	if err := s.deckRepo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a deck (soft delete) with a strategy for handling cards
func (s *DeckService) Delete(ctx context.Context, userID int64, id int64, action deck.DeleteAction, targetDeckID *int64) error {
	// 1. Find deck (validates ownership)
	existing, err := s.deckRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrDeckNotFound
	}

	// 2. Prevent deleting default deck
	if existing.GetName() == "Default" && existing.GetParentID() == nil {
		return fmt.Errorf("cannot delete the default deck")
	}

	// 3. Create pre-operation backup
	if _, err := s.backupSvc.CreatePreOperationBackup(ctx, userID); err != nil {
		return fmt.Errorf("failed to create backup before deletion: %w", err)
	}

	// 4. Handle card strategy
	var finalTargetDeckID int64
	if action == deck.ActionMoveToDefault {
		// Fetch default deck
		decks, err := s.deckRepo.FindByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to fetch user decks: %w", err)
		}
		var defaultDeck *deck.Deck
		for _, d := range decks {
			if d.GetName() == "Default" && d.IsRoot() {
				defaultDeck = d
				break
			}
		}
		if defaultDeck == nil {
			return fmt.Errorf("default deck not found")
		}
		finalTargetDeckID = defaultDeck.GetID()
	} else if action == deck.ActionMoveToDeck {
		if targetDeckID == nil {
			return fmt.Errorf("target deck ID is required for 'move_to_deck' action")
		}
		if *targetDeckID == id {
			return fmt.Errorf("cannot move cards to the deck being deleted")
		}
		// Validate target deck exists and belongs to user
		target, err := s.deckRepo.FindByID(ctx, userID, *targetDeckID)
		if err != nil {
			return fmt.Errorf("failed to validate target deck: %w", err)
		}
		if target == nil {
			return fmt.Errorf("target deck not found")
		}
		finalTargetDeckID = target.GetID()
	}

	// 4. Perform operation in transaction
	return s.tm.WithTransaction(ctx, func(txCtx context.Context) error {
		// A. Execute card action
		switch action {
		case deck.ActionDeleteCards:
			if err := s.cardRepo.DeleteByDeckRecursive(txCtx, userID, id); err != nil {
				return fmt.Errorf("failed to delete cards: %w", err)
			}
		case deck.ActionMoveToDefault, deck.ActionMoveToDeck:
			if err := s.cardRepo.MoveCards(txCtx, userID, id, finalTargetDeckID); err != nil {
				return fmt.Errorf("failed to move cards: %w", err)
			}
		}

		// B. Perform soft delete of the deck tree
		return s.deckRepo.Delete(txCtx, userID, id)
	})
}

// CreateDefaultDeck creates the initial "Default" deck for a user
func (s *DeckService) CreateDefaultDeck(ctx context.Context, userID int64) (*deck.Deck, error) {
	deckID, err := s.deckRepo.CreateDefaultDeck(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.deckRepo.FindByID(ctx, userID, deckID)
}

