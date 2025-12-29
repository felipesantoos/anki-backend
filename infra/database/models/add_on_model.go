package models

import (
	"time"
)

// AddOnModel represents the add_ons table structure in the database
type AddOnModel struct {
	ID          int64
	UserID      int64
	Code        string
	Name        string
	Version     string
	Enabled     bool
	ConfigJSON  string // JSONB stored as string
	InstalledAt time.Time
	UpdatedAt   time.Time
}

