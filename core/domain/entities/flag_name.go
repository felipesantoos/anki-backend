package entities

import (
	"time"
)

// FlagName represents a flag name entity in the domain
// It stores custom names for card flags (1-7)
type FlagName struct {
	ID          int64
	UserID      int64
	FlagNumber  int // 1-7
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsValidFlagNumber checks if the flag number is valid (1-7)
func (fn *FlagName) IsValidFlagNumber() bool {
	return fn.FlagNumber >= 1 && fn.FlagNumber <= 7
}

