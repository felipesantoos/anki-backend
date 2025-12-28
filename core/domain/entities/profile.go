package entities

import (
	"errors"
	"time"
)

var (
	// ErrAnkiWebUsernameEmpty is returned when AnkiWeb username is empty
	ErrAnkiWebUsernameEmpty = errors.New("AnkiWeb username cannot be empty")
)

// Profile represents a profile entity in the domain
// Profiles allow users to have multiple isolated collections
type Profile struct {
	ID                 int64
	UserID             int64
	Name               string
	AnkiWebSyncEnabled bool
	AnkiWebUsername    *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// IsActive checks if the profile is active (not deleted)
func (p *Profile) IsActive() bool {
	return p.DeletedAt == nil
}

// EnableAnkiWebSync enables AnkiWeb sync for the profile
func (p *Profile) EnableAnkiWebSync(username string) error {
	if username == "" {
		return ErrAnkiWebUsernameEmpty
	}

	p.AnkiWebSyncEnabled = true
	p.AnkiWebUsername = &username
	p.UpdatedAt = time.Now()
	return nil
}

// DisableAnkiWebSync disables AnkiWeb sync for the profile
func (p *Profile) DisableAnkiWebSync() {
	p.AnkiWebSyncEnabled = false
	p.AnkiWebUsername = nil
	p.UpdatedAt = time.Now()
}

