package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/app/api/dtos/request"
	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestToDomain(t *testing.T) {
	req := &request.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	email, password := ToDomain(req)
	assert.Equal(t, req.Email, email)
	assert.Equal(t, req.Password, password)
}

func TestToRegisterResponse(t *testing.T) {
	now := time.Now()
	email, _ := valueobjects.NewEmail("test@example.com")
	password, _ := valueobjects.NewPassword("password123")
	u, _ := user.NewBuilder().
		WithID(1).
		WithEmail(email).
		WithPasswordHash(password).
		WithEmailVerified(false).
		WithCreatedAt(now).
		Build()

	res := ToRegisterResponse(u)
	assert.NotNil(t, res)
	assert.Equal(t, u.GetID(), res.User.ID)
	assert.Equal(t, u.GetEmail().Value(), res.User.Email)
	assert.Equal(t, u.GetEmailVerified(), res.User.EmailVerified)
	assert.Equal(t, u.GetCreatedAt(), res.User.CreatedAt)
}

func TestToLoginResponse(t *testing.T) {
	now := time.Now()
	email, _ := valueobjects.NewEmail("test@example.com")
	password, _ := valueobjects.NewPassword("password123")
	u, _ := user.NewBuilder().
		WithID(1).
		WithEmail(email).
		WithPasswordHash(password).
		WithEmailVerified(true).
		WithCreatedAt(now).
		Build()

	accessToken := "access"
	refreshToken := "refresh"
	expiresIn := 3600

	res := ToLoginResponse(u, accessToken, refreshToken, expiresIn)
	assert.NotNil(t, res)
	assert.Equal(t, accessToken, res.AccessToken)
	assert.Equal(t, refreshToken, res.RefreshToken)
	assert.Equal(t, expiresIn, res.ExpiresIn)
	assert.Equal(t, "Bearer", res.TokenType)
	assert.Equal(t, u.GetID(), res.User.ID)
}

func TestToTokenResponse(t *testing.T) {
	accessToken := "access"
	expiresIn := 3600

	res := ToTokenResponse(accessToken, expiresIn)
	assert.NotNil(t, res)
	assert.Equal(t, accessToken, res.AccessToken)
	assert.Equal(t, expiresIn, res.ExpiresIn)
	assert.Equal(t, "Bearer", res.TokenType)
}
