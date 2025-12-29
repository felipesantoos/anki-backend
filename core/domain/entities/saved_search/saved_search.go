package savedsearch

import (
	"time"
)

// SavedSearch represents a saved search entity in the domain
// It stores user-defined search queries for reuse
type SavedSearch struct {
	id          int64
	userID      int64
	name        string
	searchQuery string
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// Getters
func (ss *SavedSearch) GetID() int64 {
	return ss.id
}

func (ss *SavedSearch) GetUserID() int64 {
	return ss.userID
}

func (ss *SavedSearch) GetName() string {
	return ss.name
}

func (ss *SavedSearch) GetSearchQuery() string {
	return ss.searchQuery
}

func (ss *SavedSearch) GetCreatedAt() time.Time {
	return ss.createdAt
}

func (ss *SavedSearch) GetUpdatedAt() time.Time {
	return ss.updatedAt
}

func (ss *SavedSearch) GetDeletedAt() *time.Time {
	return ss.deletedAt
}

// Setters
func (ss *SavedSearch) SetID(id int64) {
	ss.id = id
}

func (ss *SavedSearch) SetUserID(userID int64) {
	ss.userID = userID
}

func (ss *SavedSearch) SetName(name string) {
	ss.name = name
}

func (ss *SavedSearch) SetSearchQuery(searchQuery string) {
	ss.searchQuery = searchQuery
}

func (ss *SavedSearch) SetCreatedAt(createdAt time.Time) {
	ss.createdAt = createdAt
}

func (ss *SavedSearch) SetUpdatedAt(updatedAt time.Time) {
	ss.updatedAt = updatedAt
}

func (ss *SavedSearch) SetDeletedAt(deletedAt *time.Time) {
	ss.deletedAt = deletedAt
}

// IsActive checks if the saved search is active (not deleted)
func (ss *SavedSearch) IsActive() bool {
	return ss.deletedAt == nil
}

