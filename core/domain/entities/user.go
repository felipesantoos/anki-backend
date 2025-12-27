package entities

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// User represents a user entity in the domain
// It contains the core business logic for user management
type User struct {
	ID            int64
	Email         valueobjects.Email
	PasswordHash  valueobjects.Password
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
	DeletedAt     *time.Time
}

// IsActive checks if the user is active (not deleted)
func (u *User) IsActive() bool {
	return u.DeletedAt == nil
}

// VerifyPassword checks if the provided plain text password matches the user's password
func (u *User) VerifyPassword(plainText string) bool {
	return u.PasswordHash.Verify(plainText)
}

// MarkEmailAsVerified marks the user's email as verified
func (u *User) MarkEmailAsVerified() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}
