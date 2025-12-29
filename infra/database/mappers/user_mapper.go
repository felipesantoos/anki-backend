package mappers

import (
	"database/sql"

	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// ToDomain converts a UserModel (database representation) to a User entity (domain representation)
func ToDomain(model *models.UserModel) (*user.User, error) {
	if model == nil {
		return nil, nil
	}

	// Create email value object
	email, err := valueobjects.NewEmail(model.Email)
	if err != nil {
		return nil, err
	}

	// Create password from hash (no validation needed, already hashed)
	passwordHash := valueobjects.NewPasswordFromHash(model.PasswordHash)

	userEntity := &user.User{}
	userEntity.SetID(model.ID)
	userEntity.SetEmail(email)
	userEntity.SetPasswordHash(passwordHash)
	userEntity.SetEmailVerified(model.EmailVerified)
	userEntity.SetCreatedAt(model.CreatedAt)
	userEntity.SetUpdatedAt(model.UpdatedAt)

	// Handle nullable fields
	if model.LastLoginAt.Valid {
		userEntity.SetLastLoginAt(&model.LastLoginAt.Time)
	}

	if model.DeletedAt.Valid {
		userEntity.SetDeletedAt(&model.DeletedAt.Time)
	}

	return userEntity, nil
}

// ToModel converts a User entity (domain representation) to a UserModel (database representation)
func ToModel(userEntity *user.User) *models.UserModel {
	model := &models.UserModel{
		ID:            userEntity.GetID(),
		Email:         userEntity.GetEmail().Value(),
		PasswordHash:  userEntity.GetPasswordHash().Hash(),
		EmailVerified: userEntity.GetEmailVerified(),
		CreatedAt:     userEntity.GetCreatedAt(),
		UpdatedAt:     userEntity.GetUpdatedAt(),
	}

	// Handle nullable fields
	if userEntity.GetLastLoginAt() != nil {
		model.LastLoginAt = sql.NullTime{
			Time:  *userEntity.GetLastLoginAt(),
			Valid: true,
		}
	}

	if userEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *userEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}
