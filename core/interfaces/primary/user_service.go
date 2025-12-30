package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/user"
)

// IUserService defines the interface for general user management
type IUserService interface {
	// FindByID finds a user by ID
	FindByID(ctx context.Context, id int64) (*user.User, error)

	// Update updates user profile information
	Update(ctx context.Context, id int64, email string) (*user.User, error)

	// Delete deletes a user account (soft delete)
	Delete(ctx context.Context, id int64) error
}

