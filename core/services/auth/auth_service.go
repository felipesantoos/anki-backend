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
	"github.com/felipesantos/anki-backend/pkg/logger"
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
	accessTokenBlacklistPrefix = "access_token_blacklist"
)

// AuthService implements IAuthService
type AuthService struct {
	userRepo     secondary.IUserRepository
	deckRepo     secondary.IDeckRepository
	eventBus     secondary.IEventBus
	jwtService   *jwt.JWTService
	cacheRepo    secondary.ICacheRepository
	emailService primary.IEmailService
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo secondary.IUserRepository,
	deckRepo secondary.IDeckRepository,
	eventBus secondary.IEventBus,
	jwtService *jwt.JWTService,
	cacheRepo secondary.ICacheRepository,
	emailService primary.IEmailService,
) primary.IAuthService {
	return &AuthService{
		userRepo:     userRepo,
		deckRepo:     deckRepo,
		eventBus:     eventBus,
		jwtService:   jwtService,
		cacheRepo:    cacheRepo,
		emailService: emailService,
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

// buildAccessTokenBlacklistKey builds the Redis key for an access token blacklist entry
func buildAccessTokenBlacklistKey(token string) string {
	return fmt.Sprintf("%s:%s", accessTokenBlacklistPrefix, hashToken(token))
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

// RefreshToken generates a new access token and refresh token using a refresh token (token rotation)
// It validates the refresh token, checks if it exists in Redis, generates new tokens,
// stores the new refresh token in Redis, invalidates the old refresh token, and returns both new tokens
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*response.TokenResponse, error) {
	log := logger.GetLogger()

	// 1. Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Warn("Refresh token validation failed",
			"error", err,
		)
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// 2. Check if refresh token exists in Redis (not revoked)
	refreshTokenKey := buildRefreshTokenKey(refreshToken)
	exists, err := s.cacheRepo.Exists(ctx, refreshTokenKey)
	if err != nil {
		log.Error("Failed to check refresh token existence in Redis",
			"error", err,
			"user_id", claims.UserID,
		)
		return nil, fmt.Errorf("failed to check refresh token: %w", err)
	}
	if !exists {
		log.Warn("Refresh token not found in Redis (revoked or expired)",
			"user_id", claims.UserID,
		)
		return nil, ErrInvalidToken
	}

	// 3. Verify user still exists and is active
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		log.Error("Failed to find user during token refresh",
			"error", err,
			"user_id", claims.UserID,
		)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		log.Warn("User not found during token refresh",
			"user_id", claims.UserID,
		)
		return nil, ErrUserNotFound
	}
	if !user.IsActive() {
		log.Warn("Inactive user attempted token refresh",
			"user_id", claims.UserID,
		)
		return nil, ErrInvalidToken
	}

	// 4. Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(claims.UserID)
	if err != nil {
		log.Error("Failed to generate access token during refresh",
			"error", err,
			"user_id", claims.UserID,
		)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 5. Generate new refresh token (token rotation)
	newRefreshToken, err := s.jwtService.GenerateRefreshToken(claims.UserID)
	if err != nil {
		log.Error("Failed to generate refresh token during rotation",
			"error", err,
			"user_id", claims.UserID,
		)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 6. Store new refresh token in Redis with TTL matching refresh token expiry
	newRefreshTokenKey := buildRefreshTokenKey(newRefreshToken)
	refreshTokenTTL := s.jwtService.GetRefreshTokenExpiry()
	err = s.cacheRepo.Set(ctx, newRefreshTokenKey, fmt.Sprintf("%d", claims.UserID), refreshTokenTTL)
	if err != nil {
		log.Error("Failed to store new refresh token in Redis",
			"error", err,
			"user_id", claims.UserID,
		)
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// 7. Delete old refresh token from Redis (invalidate it - token rotation)
	err = s.cacheRepo.Delete(ctx, refreshTokenKey)
	if err != nil {
		// If deletion fails, return error to ensure token rotation integrity
		// The new token is already stored, but we need to invalidate the old one for security
		log.Error("Failed to invalidate old refresh token in Redis",
			"error", err,
			"user_id", claims.UserID,
		)
		return nil, fmt.Errorf("failed to invalidate old refresh token: %w", err)
	}

	log.Info("Token refresh successful with rotation",
		"user_id", claims.UserID,
	)

	// 8. Calculate expires_in in seconds
	expiresIn := int(s.jwtService.GetAccessTokenExpiry().Seconds())

	// 9. Build response with both new tokens
	return &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

// Logout invalidates both access token and refresh token
// It adds the access token to a blacklist in Redis and removes the refresh token
func (s *AuthService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	// 1. Invalidate access token (add to blacklist)
	if accessToken != "" {
		// Validate access token to get expiration time
		claims, err := s.jwtService.ValidateAccessToken(accessToken)
		if err == nil && claims != nil {
			// Calculate TTL: time until token expires
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				// Add access token to blacklist with TTL matching remaining expiration time
				accessTokenBlacklistKey := buildAccessTokenBlacklistKey(accessToken)
				err = s.cacheRepo.Set(ctx, accessTokenBlacklistKey, "1", ttl)
				if err != nil {
					return fmt.Errorf("failed to blacklist access token: %w", err)
				}
			}
			// If token is already expired (ttl <= 0), no need to blacklist it
		} else {
			// If token is invalid, we still try to blacklist it (idempotent operation)
			// This prevents information leakage about token validity
			// Use maximum access token expiry as TTL for invalid tokens
			accessTokenBlacklistKey := buildAccessTokenBlacklistKey(accessToken)
			maxTTL := s.jwtService.GetAccessTokenExpiry()
			_ = s.cacheRepo.Set(ctx, accessTokenBlacklistKey, "1", maxTTL)
			// Don't return error here - idempotent operation
		}
	}

	// 2. Invalidate refresh token (remove from Redis)
	if refreshToken != "" {
		// Validate refresh token (optional but good for error messages)
		_, err := s.jwtService.ValidateRefreshToken(refreshToken)
		if err != nil {
			// If token is invalid, we still try to delete it (idempotent operation)
			// This prevents information leakage about token validity
		}

		// Remove refresh token from Redis (idempotent - no error if key doesn't exist)
		refreshTokenKey := buildRefreshTokenKey(refreshToken)
		err = s.cacheRepo.Delete(ctx, refreshTokenKey)
		if err != nil {
			return fmt.Errorf("failed to delete refresh token: %w", err)
		}
	}

	return nil
}

// VerifyEmail verifies a user's email using a verification token
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// 1. Validate token using JWTService
	claims, err := s.jwtService.ValidateEmailVerificationToken(token)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// 2. Verify that token type is "email_verification" (already done in ValidateEmailVerificationToken)
	// But we can double-check for safety
	if claims.Type != "email_verification" {
		return ErrInvalidToken
	}

	// 3. Find user by ID
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}

	// 4. Check if email is already verified (idempotent operation)
	if user.EmailVerified {
		// Already verified, return success (idempotent)
		return nil
	}

	// 5. Mark email as verified
	user.MarkEmailAsVerified()

	// 6. Save to repository
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ResendVerificationEmail resends the email verification email to the user
func (s *AuthService) ResendVerificationEmail(ctx context.Context, email string) error {
	// 1. Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 2. Check if email is already verified
	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// 3. Send verification email via EmailService
	err = s.emailService.SendVerificationEmail(ctx, user.ID, user.Email.Value())
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// RequestPasswordReset generates a password reset token and sends it to the user via email
// It does not reveal if the email exists (always returns success for security)
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	// 1. Validate email format
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		// Don't return error - always return success to avoid revealing email existence
		return nil
	}

	// 2. Find user by email
	user, err := s.userRepo.FindByEmail(ctx, emailVO.Value())
	if err != nil {
		// Don't return error - always return success to avoid revealing email existence
		return nil
	}
	if user == nil {
		// User not found - return success silently (security best practice)
		return nil
	}

	// 3. Check if user is active
	if !user.IsActive() {
		// User is deleted - return success silently
		return nil
	}

	// 4. Generate password reset token
	token, err := s.jwtService.GeneratePasswordResetToken(user.ID)
	if err != nil {
		// If token generation fails, return success anyway (don't reveal failure)
		// In production, this should be logged
		return nil
	}

	// 5. Send password reset email
	err = s.emailService.SendPasswordResetEmail(ctx, user.ID, user.Email.Value(), token)
	if err != nil {
		// If email sending fails, return success anyway (don't reveal failure)
		// In production, this should be logged
		return nil
	}

	return nil
}

// ResetPassword resets a user's password using a password reset token
func (s *AuthService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	// 1. Validate token using JWTService
	claims, err := s.jwtService.ValidatePasswordResetToken(token)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// 2. Verify that token type is "password_reset" (already done in ValidatePasswordResetToken)
	// But we can double-check for safety
	if claims.Type != "password_reset" {
		return ErrInvalidToken
	}

	// 3. Find user by ID
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 4. Check if user is active
	if !user.IsActive() {
		return ErrUserNotFound
	}

	// 5. Validate new password
	passwordVO, err := valueobjects.NewPassword(newPassword)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPassword, err)
	}

	// 6. Update user password
	user.PasswordHash = passwordVO
	user.UpdatedAt = time.Now()

	// 7. Save updated user
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	// 8. Invalidate all refresh tokens for this user
	// Note: Currently, we cannot efficiently find all refresh tokens for a user
	// because the cache repository doesn't support pattern matching or scanning.
	// Refresh tokens will expire naturally based on their TTL.
	// In the future, we could:
	// - Add a method to store user->token mappings
	// - Use Redis SCAN to find all tokens for a user
	// - Store tokens in a set per user
	// For now, tokens will expire based on their TTL, and users will need to log in again
	// to get new tokens after password reset.

	return nil
}

// ChangePassword changes a user's password when authenticated
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, currentPassword string, newPassword string) error {
	// 1. Find user by ID
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 2. Check if user is active
	if !user.IsActive() {
		return ErrUserNotFound
	}

	// 3. Verify current password
	if !user.VerifyPassword(currentPassword) {
		return ErrInvalidCredentials
	}

	// 4. Validate new password
	passwordVO, err := valueobjects.NewPassword(newPassword)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPassword, err)
	}

	// 5. Update user password
	user.PasswordHash = passwordVO
	user.UpdatedAt = time.Now()

	// 6. Save updated user
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	// 7. Invalidate all refresh tokens for this user
	// Note: Currently, we cannot efficiently find all refresh tokens for a user
	// because the cache repository doesn't support pattern matching or scanning.
	// Refresh tokens will expire naturally based on their TTL.
	// In the future, we could:
	// - Add a method to store user->token mappings
	// - Use Redis SCAN to find all tokens for a user
	// - Store tokens in a set per user
	// For now, tokens will expire based on their TTL, and users will need to log in again
	// to get new tokens after password change.

	return nil
}
