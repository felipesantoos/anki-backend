package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities"
	domainEvents "github.com/felipesantos/anki-backend/core/domain/events"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/jwt"
)

var (
	// ErrEmailAlreadyExists is returned when trying to register with an existing email
	ErrEmailAlreadyExists = errors.New("email already registered")
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidPassword is returned when password doesn't meet requirements
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidCredentials is returned when email or password is incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidToken is returned when token is invalid or expired
	ErrInvalidToken = errors.New("invalid token")
)

const (
	refreshTokenKeyPrefix = "refresh_token"
)

// AuthService implements IAuthService
type AuthService struct {
	userRepo    secondary.IUserRepository
	deckRepo    secondary.IDeckRepository
	eventBus    secondary.IEventBus
	jwtService  *jwt.JWTService
	cacheRepo   secondary.ICacheRepository
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo secondary.IUserRepository,
	deckRepo secondary.IDeckRepository,
	eventBus secondary.IEventBus,
	jwtService *jwt.JWTService,
	cacheRepo secondary.ICacheRepository,
) primary.IAuthService {
	return &AuthService{
		userRepo:   userRepo,
		deckRepo:   deckRepo,
		eventBus:   eventBus,
		jwtService: jwtService,
		cacheRepo:  cacheRepo,
	}
}

// hashToken generates a SHA256 hash of the token for use as a Redis key
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// buildRefreshTokenKey builds the Redis key for a refresh token
func buildRefreshTokenKey(token string) string {
	return fmt.Sprintf("%s:%s", refreshTokenKeyPrefix, hashToken(token))
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

// Login authenticates a user and returns access and refresh tokens
func (s *AuthService) Login(ctx context.Context, email string, password string) (*response.LoginResponse, error) {
	// 1. Validate email format
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEmail, err)
	}

	// 2. Find user by email
	user, err := s.userRepo.FindByEmail(ctx, emailVO.Value())
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// 3. Check if user is active
	if !user.IsActive() {
		return nil, ErrInvalidCredentials
	}

	// 4. Verify password
	if !user.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	// 5. Update last login timestamp
	user.UpdateLastLogin()
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// 6. Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 7. Generate refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 8. Store refresh token in Redis with TTL matching refresh token expiry
	refreshTokenKey := buildRefreshTokenKey(refreshToken)
	refreshTokenTTL := s.jwtService.GetRefreshTokenExpiry()
	err = s.cacheRepo.Set(ctx, refreshTokenKey, fmt.Sprintf("%d", user.ID), refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// 9. Calculate expires_in in seconds
	expiresIn := int(s.jwtService.GetAccessTokenExpiry().Seconds())

	// 10. Build response
	return &response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User: response.UserData{
			ID:            user.ID,
			Email:         user.Email.Value(),
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
		},
	}, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*response.TokenResponse, error) {
	// 1. Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// 2. Check if token is a refresh token
	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	// 3. Check if refresh token exists in Redis (not revoked)
	refreshTokenKey := buildRefreshTokenKey(refreshToken)
	exists, err := s.cacheRepo.Exists(ctx, refreshTokenKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check refresh token: %w", err)
	}
	if !exists {
		return nil, ErrInvalidToken
	}

	// 4. Verify user still exists and is active
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if !user.IsActive() {
		return nil, ErrInvalidToken
	}

	// 5. Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 6. Calculate expires_in in seconds
	expiresIn := int(s.jwtService.GetAccessTokenExpiry().Seconds())

	// 7. Build response
	return &response.TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	}, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// 1. Validate refresh token (optional but good for error messages)
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		// If token is invalid, we still try to delete it (idempotent operation)
		// This prevents information leakage about token validity
	} else {
		// Check if token is a refresh token
		if claims.Type != "refresh" {
			return ErrInvalidToken
		}
	}

	// 2. Remove refresh token from Redis (idempotent - no error if key doesn't exist)
	refreshTokenKey := buildRefreshTokenKey(refreshToken)
	err = s.cacheRepo.Delete(ctx, refreshTokenKey)
	if err != nil {
		// Log error but don't fail logout (idempotent operation)
		// In production, you might want to log this
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}
