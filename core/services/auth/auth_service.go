package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
	domainEvents "github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

var (
	// ErrEmailAlreadyExists is returned when trying to register with an existing email
	ErrEmailAlreadyExists = errors.New("email already registered")
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidPassword is returned when password doesn't meet requirements
	ErrInvalidPassword = errors.New("invalid password")
)

// AuthService implements IAuthService
type AuthService struct {
	userRepo    secondary.IUserRepository
	deckRepo    secondary.IDeckRepository
	eventBus    secondary.IEventBus
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo secondary.IUserRepository,
	deckRepo secondary.IDeckRepository,
	eventBus secondary.IEventBus,
) primary.IAuthService {
	return &AuthService{
		userRepo: userRepo,
		deckRepo: deckRepo,
		eventBus: eventBus,
	}
}

// Register creates a new user account with email and password
// It validates the email uniqueness, hashes the password, creates the user,
// creates a default deck, and publishes a UserRegistered event
func (s *AuthService) Register(ctx context.Context, email string, password string) (*entities.User, error) {
	// 1. Validate and create email value object
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEmail, err)
	}

	// 2. Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, emailVO.Value())
	if err != nil {
		return nil, fmt.Errorf("failed to check if email exists: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// 3. Validate and create password value object (includes hashing)
	passwordVO, err := valueobjects.NewPassword(password)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPassword, err)
	}

	// 4. Create user entity
	now := time.Now()
	user := &entities.User{
		ID:            0, // Will be set after save
		Email:         emailVO,
		PasswordHash:  passwordVO,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   nil,
		DeletedAt:     nil,
	}

	// 5. Save user to database
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// 6. Create default deck for the user
	_, err = s.deckRepo.CreateDefaultDeck(ctx, user.ID)
	if err != nil {
		// If deck creation fails, we should log the error but not fail the registration
		// The user was already created, so we return success but log the deck creation failure
		// In a production system, we might want to use a transaction or a compensating action
		return nil, fmt.Errorf("failed to create default deck: %w", err)
	}

	// 7. Publish UserRegistered event
	event := &domainEvents.UserRegistered{
		UserID:    user.ID,
		Email:     user.Email.Value(),
		Timestamp: now,
	}

	if err := s.eventBus.Publish(ctx, event); err != nil {
		// Event publishing failure should not fail the registration
		// In a production system, we might want to log this or use a background job
		// For now, we'll just continue
	}

	return user, nil
}
