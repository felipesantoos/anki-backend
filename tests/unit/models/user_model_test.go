package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestUserModel_Creation(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(time.Hour)
	deletedAt := now.Add(2 * time.Hour)

	model := &models.UserModel{
		ID:            1,
		Email:         "test@example.com",
		PasswordHash:  "hash",
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   sqlNullTime(lastLogin, true),
		DeletedAt:     sqlNullTime(deletedAt, true),
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "test@example.com", model.Email)
	assert.True(t, model.EmailVerified)
	assert.True(t, model.LastLoginAt.Valid)
	assert.Equal(t, lastLogin, model.LastLoginAt.Time)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestUserModel_NullFields(t *testing.T) {
	model := &models.UserModel{
		ID:            2,
		Email:         "test2@example.com",
		LastLoginAt:   sqlNullTime(time.Time{}, false),
		DeletedAt:     sqlNullTime(time.Time{}, false),
	}

	assert.False(t, model.LastLoginAt.Valid)
	assert.False(t, model.DeletedAt.Valid)
}

