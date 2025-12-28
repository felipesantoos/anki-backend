package mappers

import (
	"database/sql"

	"github.com/felipesantos/anki-backend/core/domain/entities"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// ToDomain converts a UserModel (database representation) to a User entity (domain representation)
func ToDomain(model *models.UserModel) (*entities.User, error) {
	if model == nil {
		return nil, nil
	}

	// Create email value object
	email, err := valueobjects.NewEmail(model.Email)
	if err != nil {
		return nil, err
	}

	// Create password from hash (no validation needed, already hashed)
	password := valueobjects.NewPasswordFromHash(model.PasswordHash)

	user := &entities.User{}
	user.SetID(model.ID)
	user.SetEmail(email)
	user.SetPasswordHash(password)
	user.SetEmailVerified(model.EmailVerified)
	user.SetCreatedAt(model.CreatedAt)
	user.SetUpdatedAt(model.UpdatedAt)

	// Handle nullable fields
	if model.LastLoginAt.Valid {
		user.SetLastLoginAt(&model.LastLoginAt.Time)
	}

	if model.DeletedAt.Valid {
		user.SetDeletedAt(&model.DeletedAt.Time)
	}

	return user, nil
}

// ToModel converts a User entity (domain representation) to a UserModel (database representation)
func ToModel(user *entities.User) *models.UserModel {
	model := &models.UserModel{
		ID:            user.GetID(),
		Email:         user.GetEmail().Value(),
		PasswordHash:  user.GetPasswordHash().Hash(),
		EmailVerified: user.GetEmailVerified(),
		CreatedAt:     user.GetCreatedAt(),
		UpdatedAt:     user.GetUpdatedAt(),
	}

	// Handle nullable fields
	if user.GetLastLoginAt() != nil {
		model.LastLoginAt = sql.NullTime{
			Time:  *user.GetLastLoginAt(),
			Valid: true,
		}
	}

	if user.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *user.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}
