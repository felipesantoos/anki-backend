package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestUserToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(time.Hour)
	deletedAt := now.Add(2 * time.Hour)

	model := &models.UserModel{
		ID:            1,
		Email:         "test@example.com",
		PasswordHash:  "hashed_password",
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   sqlNullTime(lastLogin, true),
		DeletedAt:     sqlNullTime(deletedAt, true),
	}

	entity, err := UserToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, "test@example.com", entity.GetEmail().Value())
	assert.Equal(t, "hashed_password", entity.GetPasswordHash().Hash())
	assert.True(t, entity.GetEmailVerified())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now, entity.GetUpdatedAt())
	assert.NotNil(t, entity.GetLastLoginAt())
	assert.Equal(t, lastLogin, *entity.GetLastLoginAt())
	assert.NotNil(t, entity.GetDeletedAt())
	assert.Equal(t, deletedAt, *entity.GetDeletedAt())
}

func TestUserToDomain_WithNullFields(t *testing.T) {
	now := time.Now()

	model := &models.UserModel{
		ID:            2,
		Email:         "user2@example.com",
		PasswordHash:  "other_hash",
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   sqlNullTime(time.Time{}, false),
		DeletedAt:     sqlNullTime(time.Time{}, false),
	}

	entity, err := UserToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetLastLoginAt())
	assert.Nil(t, entity.GetDeletedAt())
}

func TestUserToDomain_NilInput(t *testing.T) {
	entity, err := UserToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestUserToDomain_InvalidEmail(t *testing.T) {
	model := &models.UserModel{
		ID:    1,
		Email: "invalid-email",
	}

	entity, err := UserToDomain(model)
	assert.Error(t, err)
	assert.Nil(t, entity)
}

func TestUserToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(time.Hour)
	deletedAt := now.Add(2 * time.Hour)

	email, _ := valueobjects.NewEmail("test@example.com")
	passwordHash := valueobjects.NewPasswordFromHash("hashed_password")

	entity, _ := UserToDomain(&models.UserModel{
		ID:            1,
		Email:         "test@example.com",
		PasswordHash:  "hashed_password",
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   sqlNullTime(lastLogin, true),
		DeletedAt:     sqlNullTime(deletedAt, true),
	})

	model := UserToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, email.Value(), model.Email)
	assert.Equal(t, passwordHash.Hash(), model.PasswordHash)
	assert.True(t, model.EmailVerified)
	assert.True(t, model.LastLoginAt.Valid)
	assert.Equal(t, lastLogin, model.LastLoginAt.Time)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestUserToModel_WithNullFields(t *testing.T) {
	entity, _ := UserToDomain(&models.UserModel{
		ID:            2,
		Email:         "user2@example.com",
		PasswordHash:  "other_hash",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastLoginAt:   sqlNullTime(time.Time{}, false),
		DeletedAt:     sqlNullTime(time.Time{}, false),
	})

	model := UserToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.LastLoginAt.Valid)
	assert.False(t, model.DeletedAt.Valid)
}

