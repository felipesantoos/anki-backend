package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities"
)

// IUserRepository defines the interface for user data persistence
type IUserRepository interface {
	// Save saves or updates a user in the database
	// If the user has an ID, it updates the existing user
	// If the user has no ID, it creates a new user and returns it with the ID set
	Save(ctx context.Context, user *entities.User) error

	// FindByEmail finds a user by email address
	// Returns the user if found, nil if not found, or an error if the query fails
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindByID finds a user by ID
	// Returns the user if found, nil if not found, or an error if the query fails
	FindByID(ctx context.Context, id int64) (*entities.User, error)

	// ExistsByEmail checks if a user with the given email already exists
	// Returns true if exists, false if not, or an error if the query fails
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
