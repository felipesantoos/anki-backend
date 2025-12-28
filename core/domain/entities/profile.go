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
	id                 int64
	userID             int64
	name               string
	ankiWebSyncEnabled bool
	ankiWebUsername    *string
	createdAt          time.Time
	updatedAt          time.Time
	deletedAt          *time.Time
}

// Getters
func (p *Profile) GetID() int64 {
	return p.id
}

func (p *Profile) GetUserID() int64 {
	return p.userID
}

func (p *Profile) GetName() string {
	return p.name
}

func (p *Profile) GetAnkiWebSyncEnabled() bool {
	return p.ankiWebSyncEnabled
}

func (p *Profile) GetAnkiWebUsername() *string {
	return p.ankiWebUsername
}

func (p *Profile) GetCreatedAt() time.Time {
	return p.createdAt
}

func (p *Profile) GetUpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Profile) GetDeletedAt() *time.Time {
	return p.deletedAt
}

// Setters
func (p *Profile) SetID(id int64) {
	p.id = id
}

func (p *Profile) SetUserID(userID int64) {
	p.userID = userID
}

func (p *Profile) SetName(name string) {
	p.name = name
}

func (p *Profile) SetAnkiWebSyncEnabled(ankiWebSyncEnabled bool) {
	p.ankiWebSyncEnabled = ankiWebSyncEnabled
}

func (p *Profile) SetAnkiWebUsername(ankiWebUsername *string) {
	p.ankiWebUsername = ankiWebUsername
}

func (p *Profile) SetCreatedAt(createdAt time.Time) {
	p.createdAt = createdAt
}

func (p *Profile) SetUpdatedAt(updatedAt time.Time) {
	p.updatedAt = updatedAt
}

func (p *Profile) SetDeletedAt(deletedAt *time.Time) {
	p.deletedAt = deletedAt
}

// IsActive checks if the profile is active (not deleted)
func (p *Profile) IsActive() bool {
	return p.deletedAt == nil
}

// EnableAnkiWebSync enables AnkiWeb sync for the profile
func (p *Profile) EnableAnkiWebSync(username string) error {
	if username == "" {
		return ErrAnkiWebUsernameEmpty
	}

	p.ankiWebSyncEnabled = true
	p.ankiWebUsername = &username
	p.updatedAt = time.Now()
	return nil
}

// DisableAnkiWebSync disables AnkiWeb sync for the profile
func (p *Profile) DisableAnkiWebSync() {
	p.ankiWebSyncEnabled = false
	p.ankiWebUsername = nil
	p.updatedAt = time.Now()
}

