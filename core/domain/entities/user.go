package entities

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// User represents a user entity in the domain
// It contains the core business logic for user management
type User struct {
	id            int64
	email         valueobjects.Email
	passwordHash  valueobjects.Password
	emailVerified bool
	createdAt     time.Time
	updatedAt     time.Time
	lastLoginAt   *time.Time
	deletedAt     *time.Time
}

// Getters
func (u *User) GetID() int64 {
	return u.id
}

func (u *User) GetEmail() valueobjects.Email {
	return u.email
}

func (u *User) GetPasswordHash() valueobjects.Password {
	return u.passwordHash
}

func (u *User) GetEmailVerified() bool {
	return u.emailVerified
}

func (u *User) GetCreatedAt() time.Time {
	return u.createdAt
}

func (u *User) GetUpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) GetLastLoginAt() *time.Time {
	return u.lastLoginAt
}

func (u *User) GetDeletedAt() *time.Time {
	return u.deletedAt
}

// Setters
func (u *User) SetID(id int64) {
	u.id = id
}

func (u *User) SetEmail(email valueobjects.Email) {
	u.email = email
}

func (u *User) SetPasswordHash(passwordHash valueobjects.Password) {
	u.passwordHash = passwordHash
}

func (u *User) SetEmailVerified(emailVerified bool) {
	u.emailVerified = emailVerified
}

func (u *User) SetCreatedAt(createdAt time.Time) {
	u.createdAt = createdAt
}

func (u *User) SetUpdatedAt(updatedAt time.Time) {
	u.updatedAt = updatedAt
}

func (u *User) SetLastLoginAt(lastLoginAt *time.Time) {
	u.lastLoginAt = lastLoginAt
}

func (u *User) SetDeletedAt(deletedAt *time.Time) {
	u.deletedAt = deletedAt
}

// IsActive checks if the user is active (not deleted)
func (u *User) IsActive() bool {
	return u.deletedAt == nil
}

// VerifyPassword checks if the provided plain text password matches the user's password
func (u *User) VerifyPassword(plainText string) bool {
	return u.passwordHash.Verify(plainText)
}

// MarkEmailAsVerified marks the user's email as verified
func (u *User) MarkEmailAsVerified() {
	u.emailVerified = true
	u.updatedAt = time.Now()
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.lastLoginAt = &now
	u.updatedAt = now
}
