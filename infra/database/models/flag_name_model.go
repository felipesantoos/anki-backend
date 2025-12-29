package models

import (
	"time"
)

// FlagNameModel represents the flag_names table structure in the database
type FlagNameModel struct {
	ID         int64
	UserID     int64
	FlagNumber int
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

