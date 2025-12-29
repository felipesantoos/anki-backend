package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// UserRepository implements IUserRepository using PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *sql.DB) secondary.IUserRepository {
	return &UserRepository{
		db: db,
	}
}

// Save saves or updates a user in the database
// If the user has an ID, it updates the existing user
// If the user has no ID, it creates a new user and returns it with the ID set
func (r *UserRepository) Save(ctx context.Context, userEntity *user.User) error {
	if userEntity.GetID() == 0 {
		// Insert new user
		query := `
			INSERT INTO users (email, password_hash, email_verified, created_at, updated_at, last_login_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`

		model := mappers.ToModel(userEntity)
		var lastLoginAt, deletedAt sql.NullTime

		if model.LastLoginAt.Valid {
			lastLoginAt = model.LastLoginAt
		}
		if model.DeletedAt.Valid {
			deletedAt = model.DeletedAt
		}

		var userID int64
		err := r.db.QueryRowContext(ctx, query,
			model.Email,
			model.PasswordHash,
			model.EmailVerified,
			model.CreatedAt,
			model.UpdatedAt,
			lastLoginAt,
			deletedAt,
		).Scan(&userID)
		if err == nil {
			userEntity.SetID(userID)
		}

		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	}

	// Update existing user
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, email_verified = $3, updated_at = $4, last_login_at = $5, deleted_at = $6
		WHERE id = $7
	`

	model := mappers.ToModel(userEntity)
	var lastLoginAt, deletedAt sql.NullTime

	if model.LastLoginAt.Valid {
		lastLoginAt = model.LastLoginAt
	}
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Email,
		model.PasswordHash,
		model.EmailVerified,
		model.UpdatedAt,
		lastLoginAt,
		deletedAt,
		model.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Update updates an existing user in the database
func (r *UserRepository) Update(ctx context.Context, userEntity *user.User) error {
	// Use Save which handles both create and update
	return r.Save(ctx, userEntity)
}

// FindByEmail finds a user by email address
// Returns the user if found, nil if not found, or an error if the query fails
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, email_verified, created_at, updated_at, last_login_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var model models.UserModel
	var lastLoginAt, deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.EmailVerified,
		&model.CreatedAt,
		&model.UpdatedAt,
		&lastLoginAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	model.LastLoginAt = lastLoginAt
	model.DeletedAt = deletedAt

	return mappers.ToDomain(&model)
}

// FindByID finds a user by ID
// Returns the user if found, nil if not found, or an error if the query fails
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, email_verified, created_at, updated_at, last_login_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var model models.UserModel
	var lastLoginAt, deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.EmailVerified,
		&model.CreatedAt,
		&model.UpdatedAt,
		&lastLoginAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	model.LastLoginAt = lastLoginAt
	model.DeletedAt = deletedAt

	return mappers.ToDomain(&model)
}

// ExistsByEmail checks if a user with the given email already exists
// Returns true if exists, false if not, or an error if the query fails
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// Ensure UserRepository implements IUserRepository
var _ secondary.IUserRepository = (*UserRepository)(nil)
