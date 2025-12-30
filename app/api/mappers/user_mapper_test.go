package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestToUserResponse(t *testing.T) {
	now := time.Now()
	email, _ := valueobjects.NewEmail("user@example.com")
	password, _ := valueobjects.NewPassword("password123")
	
	u, _ := user.NewBuilder().
		WithID(10).
		WithEmail(email).
		WithPasswordHash(password).
		WithEmailVerified(true).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToUserResponse(u)
		assert.NotNil(t, res)
		assert.Equal(t, u.GetID(), res.ID)
		assert.Equal(t, u.GetEmail().Value(), res.Email)
		assert.Equal(t, u.GetEmailVerified(), res.EmailVerified)
		assert.Equal(t, u.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, u.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToUserResponse(nil)
		assert.Nil(t, res)
	})
}
