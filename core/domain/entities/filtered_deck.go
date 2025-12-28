package entities

import (
	"time"
)

// FilteredDeck represents a filtered deck entity in the domain
// Filtered decks are dynamically generated based on search criteria
type FilteredDeck struct {
	ID            int64
	UserID        int64
	Name          string
	SearchFilter  string
	SecondFilter  *string
	LimitCards    int
	OrderBy       string
	Reschedule    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastRebuildAt *time.Time
	DeletedAt     *time.Time
}

// IsActive checks if the filtered deck is active (not deleted)
func (fd *FilteredDeck) IsActive() bool {
	return fd.DeletedAt == nil
}

// NeedsRebuild checks if the filtered deck needs to be rebuilt
// This is a domain method - actual rebuild logic should be in service layer
func (fd *FilteredDeck) NeedsRebuild() bool {
	return fd.IsActive() && (fd.LastRebuildAt == nil || fd.UpdatedAt.After(*fd.LastRebuildAt))
}

