package entities

import (
	"time"
)

// AddOn represents an add-on entity in the domain
// It stores information about user-installed add-ons
type AddOn struct {
	ID          int64
	UserID      int64
	Code        string
	Name        string
	Version     string
	Enabled     bool
	ConfigJSON  string // JSONB in database
	InstalledAt time.Time
	UpdatedAt   time.Time
}

// IsEnabled checks if the add-on is enabled
func (ao *AddOn) IsEnabled() bool {
	return ao.Enabled
}

// Enable enables the add-on
func (ao *AddOn) Enable() {
	if !ao.Enabled {
		ao.Enabled = true
		ao.UpdatedAt = time.Now()
	}
}

// Disable disables the add-on
func (ao *AddOn) Disable() {
	if ao.Enabled {
		ao.Enabled = false
		ao.UpdatedAt = time.Now()
	}
}

