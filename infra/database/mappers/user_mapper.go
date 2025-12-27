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

	user := &entities.User{
		ID:            model.ID,
		Email:         email,
		PasswordHash:  password,
		EmailVerified: model.EmailVerified,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}

	// Handle nullable fields
	if model.LastLoginAt.Valid {
		user.LastLoginAt = &model.LastLoginAt.Time
	}

	if model.DeletedAt.Valid {
		user.DeletedAt = &model.DeletedAt.Time
	}

	return user, nil
}

// ToModel converts a User entity (domain representation) to a UserModel (database representation)
func ToModel(user *entities.User) *models.UserModel {
	model := &models.UserModel{
		ID:            user.ID,
		Email:         user.Email.Value(),
		PasswordHash:  user.PasswordHash.Hash(),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	// Handle nullable fields
	if user.LastLoginAt != nil {
		model.LastLoginAt = sql.NullTime{
			Time:  *user.LastLoginAt,
			Valid: true,
		}
	}

	if user.DeletedAt != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *user.DeletedAt,
			Valid: true,
		}
	}

	return model
}
