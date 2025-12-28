package entities

import (
	"time"
)

// FlagName represents a flag name entity in the domain
// It stores custom names for card flags (1-7)
type FlagName struct {
	id          int64
	userID      int64
	flagNumber  int // 1-7
	name        string
	createdAt   time.Time
	updatedAt   time.Time
}

// Getters
func (fn *FlagName) GetID() int64 {
	return fn.id
}

func (fn *FlagName) GetUserID() int64 {
	return fn.userID
}

func (fn *FlagName) GetFlagNumber() int {
	return fn.flagNumber
}

func (fn *FlagName) GetName() string {
	return fn.name
}

func (fn *FlagName) GetCreatedAt() time.Time {
	return fn.createdAt
}

func (fn *FlagName) GetUpdatedAt() time.Time {
	return fn.updatedAt
}

// Setters
func (fn *FlagName) SetID(id int64) {
	fn.id = id
}

func (fn *FlagName) SetUserID(userID int64) {
	fn.userID = userID
}

func (fn *FlagName) SetFlagNumber(flagNumber int) {
	fn.flagNumber = flagNumber
}

func (fn *FlagName) SetName(name string) {
	fn.name = name
}

func (fn *FlagName) SetCreatedAt(createdAt time.Time) {
	fn.createdAt = createdAt
}

func (fn *FlagName) SetUpdatedAt(updatedAt time.Time) {
	fn.updatedAt = updatedAt
}

// IsValidFlagNumber checks if the flag number is valid (1-7)
func (fn *FlagName) IsValidFlagNumber() bool {
	return fn.flagNumber >= 1 && fn.flagNumber <= 7
}

