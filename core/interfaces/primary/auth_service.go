package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
)

// IAuthService defines the interface for authentication operations
type IAuthService interface {
	// Register creates a new user account with email and password
	// It validates the email uniqueness, hashes the password, creates the user,
	// creates a default deck, and publishes a UserRegistered event
	// Returns the created user or an error if registration fails
	Register(ctx context.Context, email string, password string) (*user.User, error)

	// Login authenticates a user and returns access and refresh tokens
	// It validates credentials, generates JWT tokens, stores refresh token in Redis,
	// creates a session with metadata (IP, user agent), and updates the user's last login timestamp
	// Returns login response with tokens and user data, or an error if authentication fails
	Login(ctx context.Context, email string, password string, ipAddress string, userAgent string) (*response.LoginResponse, error)

	// RefreshToken generates a new access token and refresh token using a refresh token (token rotation)
	// It validates the refresh token, checks if it exists in Redis, generates new tokens,
	// stores the new refresh token in Redis, invalidates the old refresh token, and returns both new tokens
	// Returns token response with new access token and new refresh token, or an error if refresh token is invalid
	RefreshToken(ctx context.Context, refreshToken string) (*response.TokenResponse, error)

	// Logout invalidates both access token and refresh token
	// It adds the access token to a blacklist in Redis and removes the refresh token
	// Either accessToken or refreshToken (or both) should be provided
	// Returns an error if the operation fails
	Logout(ctx context.Context, accessToken string, refreshToken string) error

	// VerifyEmail verifies a user's email using a verification token
	// It validates the token, checks if it's an email verification token,
	// and marks the user's email as verified (idempotent operation)
	// Returns an error if the token is invalid, expired, or verification fails
	VerifyEmail(ctx context.Context, token string) error

	// ResendVerificationEmail resends the email verification email to the user
	// It checks if the email is already verified and returns an error if it is
	// Returns an error if the user is not found or email sending fails
	ResendVerificationEmail(ctx context.Context, email string) error

	// RequestPasswordReset generates a password reset token and sends it to the user via email
	// It does not reveal if the email exists (always returns success for security)
	// If the email exists, it generates a token and sends the reset email
	// Returns an error only if email sending fails (but user existence is never revealed)
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPassword resets a user's password using a password reset token
	// It validates the token, checks if it's a password reset token,
	// validates the new password, updates the user's password,
	// and invalidates all refresh tokens for the user
	// Returns an error if the token is invalid, expired, or password reset fails
	ResetPassword(ctx context.Context, token string, newPassword string) error

	// ChangePassword changes a user's password when authenticated
	// It validates the current password, validates the new password,
	// updates the user's password, and invalidates all refresh tokens for the user
	// Returns an error if the current password is incorrect, new password is invalid, or update fails
	ChangePassword(ctx context.Context, userID int64, currentPassword string, newPassword string) error
}
