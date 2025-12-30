package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// ProfileService implements IProfileService
type ProfileService struct {
	profileRepo secondary.IProfileRepository
}

// NewProfileService creates a new ProfileService instance
func NewProfileService(profileRepo secondary.IProfileRepository) primary.IProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
	}
}

// Create creates a new profile for a user
func (s *ProfileService) Create(ctx context.Context, userID int64, name string) (*profile.Profile, error) {
	// 1. Check if profile with same name exists for user
	found, err := s.profileRepo.FindByName(ctx, userID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check profile existence: %w", err)
	}
	if found != nil {
		return nil, fmt.Errorf("profile with name %s already exists", name)
	}

	// 2. Create profile entity using builder
	now := time.Now()
	profileEntity, err := profile.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName(name).
		WithAnkiWebSyncEnabled(false).
		WithAnkiWebUsername(nil).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build profile entity: %w", err)
	}

	// 3. Save to repository
	if err := s.profileRepo.Save(ctx, userID, profileEntity); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profileEntity, nil
}

// FindByID finds a profile by ID
func (s *ProfileService) FindByID(ctx context.Context, userID int64, id int64) (*profile.Profile, error) {
	return s.profileRepo.FindByID(ctx, userID, id)
}

// FindByUserID finds all profiles for a user
func (s *ProfileService) FindByUserID(ctx context.Context, userID int64) ([]*profile.Profile, error) {
	return s.profileRepo.FindByUserID(ctx, userID)
}

// Update updates an existing profile
func (s *ProfileService) Update(ctx context.Context, userID int64, id int64, name string) (*profile.Profile, error) {
	// 1. Find existing profile
	existing, err := s.profileRepo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("profile not found")
	}

	// 2. If name changed, check for conflicts
	if existing.GetName() != name {
		found, err := s.profileRepo.FindByName(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return nil, fmt.Errorf("profile with name %s already exists", name)
		}
	}

	// 3. Update entity
	existing.SetName(name)
	existing.SetUpdatedAt(time.Now())

	// 4. Save
	if err := s.profileRepo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a profile (soft delete)
func (s *ProfileService) Delete(ctx context.Context, userID int64, id int64) error {
	// 1. Check if profile exists
	existing, err := s.profileRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("profile not found")
	}

	// 2. Perform soft delete
	return s.profileRepo.Delete(ctx, userID, id)
}

// EnableSync enables AnkiWeb sync for a profile
func (s *ProfileService) EnableSync(ctx context.Context, userID int64, id int64, username string) error {
	// 1. Find profile
	existing, err := s.profileRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("profile not found")
	}

	// 2. Enable sync in entity
	if err := existing.EnableAnkiWebSync(username); err != nil {
		return err
	}

	// 3. Save
	return s.profileRepo.Update(ctx, userID, id, existing)
}

// DisableSync disables AnkiWeb sync for a profile
func (s *ProfileService) DisableSync(ctx context.Context, userID int64, id int64) error {
	// 1. Find profile
	existing, err := s.profileRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("profile not found")
	}

	// 2. Disable sync in entity
	existing.DisableAnkiWebSync()

	// 3. Save
	return s.profileRepo.Update(ctx, userID, id, existing)
}

