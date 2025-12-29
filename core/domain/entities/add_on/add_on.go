package addon

import (
	"time"
)

// AddOn represents an add-on entity in the domain
// It stores information about user-installed add-ons
type AddOn struct {
	id          int64
	userID      int64
	code        string
	name        string
	version     string
	enabled     bool
	configJSON  string // JSONB in database
	installedAt time.Time
	updatedAt   time.Time
}

// Getters
func (ao *AddOn) GetID() int64 {
	return ao.id
}

func (ao *AddOn) GetUserID() int64 {
	return ao.userID
}

func (ao *AddOn) GetCode() string {
	return ao.code
}

func (ao *AddOn) GetName() string {
	return ao.name
}

func (ao *AddOn) GetVersion() string {
	return ao.version
}

func (ao *AddOn) GetEnabled() bool {
	return ao.enabled
}

func (ao *AddOn) GetConfigJSON() string {
	return ao.configJSON
}

func (ao *AddOn) GetInstalledAt() time.Time {
	return ao.installedAt
}

func (ao *AddOn) GetUpdatedAt() time.Time {
	return ao.updatedAt
}

// Setters
func (ao *AddOn) SetID(id int64) {
	ao.id = id
}

func (ao *AddOn) SetUserID(userID int64) {
	ao.userID = userID
}

func (ao *AddOn) SetCode(code string) {
	ao.code = code
}

func (ao *AddOn) SetName(name string) {
	ao.name = name
}

func (ao *AddOn) SetVersion(version string) {
	ao.version = version
}

func (ao *AddOn) SetEnabled(enabled bool) {
	ao.enabled = enabled
}

func (ao *AddOn) SetConfigJSON(configJSON string) {
	ao.configJSON = configJSON
}

func (ao *AddOn) SetInstalledAt(installedAt time.Time) {
	ao.installedAt = installedAt
}

func (ao *AddOn) SetUpdatedAt(updatedAt time.Time) {
	ao.updatedAt = updatedAt
}

// IsEnabled checks if the add-on is enabled
func (ao *AddOn) IsEnabled() bool {
	return ao.enabled
}

// Enable enables the add-on
func (ao *AddOn) Enable() {
	if !ao.enabled {
		ao.enabled = true
		ao.updatedAt = time.Now()
	}
}

// Disable disables the add-on
func (ao *AddOn) Disable() {
	if ao.enabled {
		ao.enabled = false
		ao.updatedAt = time.Now()
	}
}

