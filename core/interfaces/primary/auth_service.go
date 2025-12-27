package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities"
)

// IAuthService defines the interface for authentication operations
type IAuthService interface {
	// Register creates a new user account with email and password
	// It validates the email uniqueness, hashes the password, creates the user,
	// creates a default deck, and publishes a UserRegistered event
	// Returns the created user or an error if registration fails
	Register(ctx context.Context, email string, password string) (*entities.User, error)

	// Login authenticates a user and returns access and refresh tokens
	// It validates credentials, generates JWT tokens, stores refresh token in Redis,
	// and updates the user's last login timestamp
	// Returns login response with tokens and user data, or an error if authentication fails
	Login(ctx context.Context, email string, password string) (*response.LoginResponse, error)

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
}
