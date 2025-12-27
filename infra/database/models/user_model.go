package models

import (
	"database/sql"
	"time"
)

// UserModel represents the user table structure in the database
// It uses database/sql nullable types for optional fields
type UserModel struct {
	ID            int64
	Email         string
	PasswordHash  string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   sql.NullTime
	DeletedAt     sql.NullTime
}
