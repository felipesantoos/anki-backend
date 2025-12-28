package entities

import (
	"time"
)

// SavedSearch represents a saved search entity in the domain
// It stores user-defined search queries for reuse
type SavedSearch struct {
	ID          int64
	UserID      int64
	Name        string
	SearchQuery string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// IsActive checks if the saved search is active (not deleted)
func (ss *SavedSearch) IsActive() bool {
	return ss.DeletedAt == nil
}

