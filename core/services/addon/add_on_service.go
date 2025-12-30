package addon

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// AddOnService implements IAddOnService
type AddOnService struct {
	repo secondary.IAddOnRepository
}

// NewAddOnService creates a new AddOnService instance
func NewAddOnService(repo secondary.IAddOnRepository) primary.IAddOnService {
	return &AddOnService{
		repo: repo,
	}
}

// Install records a new add-on installation
func (s *AddOnService) Install(ctx context.Context, userID int64, code string, name string, version string, configJSON string) (*addon.AddOn, error) {
	existing, err := s.repo.FindByCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing != nil {
		// Update existing installation
		existing.SetName(name)
		existing.SetVersion(version)
		existing.SetConfigJSON(configJSON)
		existing.SetUpdatedAt(now)
		if err := s.repo.Update(ctx, userID, existing.GetID(), existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new installation
	newAddOn, err := addon.NewBuilder().
		WithUserID(userID).
		WithCode(code).
		WithName(name).
		WithVersion(version).
		WithEnabled(true).
		WithConfigJSON(configJSON).
		WithInstalledAt(now).
		WithUpdatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, newAddOn); err != nil {
		return nil, err
	}

	return newAddOn, nil
}

// FindByUserID finds all add-ons for a user
func (s *AddOnService) FindByUserID(ctx context.Context, userID int64) ([]*addon.AddOn, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// UpdateConfig updates an add-on's configuration
func (s *AddOnService) UpdateConfig(ctx context.Context, userID int64, code string, configJSON string) (*addon.AddOn, error) {
	existing, err := s.repo.FindByCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("add-on not found")
	}

	existing.SetConfigJSON(configJSON)
	existing.SetUpdatedAt(time.Now())

	if err := s.repo.Update(ctx, userID, existing.GetID(), existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// ToggleEnabled enables or disables an add-on
func (s *AddOnService) ToggleEnabled(ctx context.Context, userID int64, code string, enabled bool) (*addon.AddOn, error) {
	existing, err := s.repo.FindByCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("add-on not found")
	}

	existing.SetEnabled(enabled)
	existing.SetUpdatedAt(time.Now())

	if err := s.repo.Update(ctx, userID, existing.GetID(), existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Uninstall removes an add-on
func (s *AddOnService) Uninstall(ctx context.Context, userID int64, code string) error {
	existing, err := s.repo.FindByCode(ctx, userID, code)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("add-on not found")
	}

	return s.repo.Delete(ctx, userID, existing.GetID())
}

