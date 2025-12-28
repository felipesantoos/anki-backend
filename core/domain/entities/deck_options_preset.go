package entities

import (
	"time"
)

// DeckOptionsPreset represents a deck options preset entity in the domain
// It stores reusable deck configuration presets
type DeckOptionsPreset struct {
	id          int64
	userID      int64
	name        string
	optionsJSON string // JSONB in database
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// Getters
func (dop *DeckOptionsPreset) GetID() int64 {
	return dop.id
}

func (dop *DeckOptionsPreset) GetUserID() int64 {
	return dop.userID
}

func (dop *DeckOptionsPreset) GetName() string {
	return dop.name
}

func (dop *DeckOptionsPreset) GetOptionsJSON() string {
	return dop.optionsJSON
}

func (dop *DeckOptionsPreset) GetCreatedAt() time.Time {
	return dop.createdAt
}

func (dop *DeckOptionsPreset) GetUpdatedAt() time.Time {
	return dop.updatedAt
}

func (dop *DeckOptionsPreset) GetDeletedAt() *time.Time {
	return dop.deletedAt
}

// Setters
func (dop *DeckOptionsPreset) SetID(id int64) {
	dop.id = id
}

func (dop *DeckOptionsPreset) SetUserID(userID int64) {
	dop.userID = userID
}

func (dop *DeckOptionsPreset) SetName(name string) {
	dop.name = name
}

func (dop *DeckOptionsPreset) SetOptionsJSON(optionsJSON string) {
	dop.optionsJSON = optionsJSON
}

func (dop *DeckOptionsPreset) SetCreatedAt(createdAt time.Time) {
	dop.createdAt = createdAt
}

func (dop *DeckOptionsPreset) SetUpdatedAt(updatedAt time.Time) {
	dop.updatedAt = updatedAt
}

func (dop *DeckOptionsPreset) SetDeletedAt(deletedAt *time.Time) {
	dop.deletedAt = deletedAt
}

// IsActive checks if the preset is active (not deleted)
func (dop *DeckOptionsPreset) IsActive() bool {
	return dop.deletedAt == nil
}

