package deck

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// DeckOptionsPresetService implements IDeckOptionsPresetService
type DeckOptionsPresetService struct {
	repo secondary.IDeckOptionsPresetRepository
}

// NewDeckOptionsPresetService creates a new DeckOptionsPresetService instance
func NewDeckOptionsPresetService(repo secondary.IDeckOptionsPresetRepository) primary.IDeckOptionsPresetService {
	return &DeckOptionsPresetService{
		repo: repo,
	}
}

// Create creates a new options preset
func (s *DeckOptionsPresetService) Create(ctx context.Context, userID int64, name string, optionsJSON string) (*deckoptionspreset.DeckOptionsPreset, error) {
	exists, err := s.repo.FindByName(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, fmt.Errorf("preset with name %s already exists", name)
	}

	now := time.Now()
	p, err := deckoptionspreset.NewBuilder().
		WithUserID(userID).
		WithName(name).
		WithOptionsJSON(optionsJSON).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, p); err != nil {
		return nil, err
	}

	return p, nil
}

// FindByUserID finds all presets for a user
func (s *DeckOptionsPresetService) FindByUserID(ctx context.Context, userID int64) ([]*deckoptionspreset.DeckOptionsPreset, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// Update updates an existing preset
func (s *DeckOptionsPresetService) Update(ctx context.Context, userID int64, id int64, name string, optionsJSON string) (*deckoptionspreset.DeckOptionsPreset, error) {
	existing, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("preset not found")
	}

	if existing.GetName() != name {
		conflict, err := s.repo.FindByName(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if conflict != nil {
			return nil, fmt.Errorf("preset with name %s already exists", name)
		}
	}

	existing.SetName(name)
	existing.SetOptionsJSON(optionsJSON)
	existing.SetUpdatedAt(time.Now())

	if err := s.repo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a preset
func (s *DeckOptionsPresetService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

