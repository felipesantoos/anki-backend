package user

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// UserService implements IUserService
type UserService struct {
	userRepo secondary.IUserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo secondary.IUserRepository) primary.IUserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// FindByID finds a user by ID
func (s *UserService) FindByID(ctx context.Context, id int64) (*user.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// Update updates user profile information
func (s *UserService) Update(ctx context.Context, id int64, email string) (*user.User, error) {
	existing, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("user not found")
	}

	if existing.GetEmail().Value() != email {
		emailVO, err := valueobjects.NewEmail(email)
		if err != nil {
			return nil, err
		}

		exists, err := s.userRepo.ExistsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("email %s already in use", email)
		}

		existing.SetEmail(emailVO)
		existing.SetEmailVerified(false)
	}

	existing.SetUpdatedAt(time.Now())

	if err := s.userRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a user account (soft delete)
func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.userRepo.Delete(ctx, id)
}

