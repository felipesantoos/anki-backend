package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities"
)

// IAuthService defines the interface for authentication operations
type IAuthService interface {
	// Register creates a new user account with email and password
	// It validates the email uniqueness, hashes the password, creates the user,
	// creates a default deck, and publishes a UserRegistered event
	// Returns the created user or an error if registration fails
	Register(ctx context.Context, email string, password string) (*entities.User, error)
}
