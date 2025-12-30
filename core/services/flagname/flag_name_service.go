package flagname

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// FlagNameService implements IFlagNameService
type FlagNameService struct {
	flagRepo secondary.IFlagNameRepository
}

// NewFlagNameService creates a new FlagNameService instance
func NewFlagNameService(flagRepo secondary.IFlagNameRepository) primary.IFlagNameService {
	return &FlagNameService{
		flagRepo: flagRepo,
	}
}

// FindByUserID finds all flag names for a user
func (s *FlagNameService) FindByUserID(ctx context.Context, userID int64) ([]*flagname.FlagName, error) {
	return s.flagRepo.FindByUserID(ctx, userID)
}

// Update updates a flag name
func (s *FlagNameService) Update(ctx context.Context, userID int64, flagNumber int, name string) (*flagname.FlagName, error) {
	// 1. Find existing flag name
	existing, err := s.flagRepo.FindByFlagNumber(ctx, userID, flagNumber)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing == nil {
		// Create new if not exists
		existing, err = flagname.NewBuilder().
			WithUserID(userID).
			WithFlagNumber(flagNumber).
			WithName(name).
			WithCreatedAt(now).
			WithUpdatedAt(now).
			Build()
		if err != nil {
			return nil, err
		}
	} else {
		// Update existing
		existing.SetName(name)
		existing.SetUpdatedAt(now)
	}

	if err := s.flagRepo.Save(ctx, userID, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

