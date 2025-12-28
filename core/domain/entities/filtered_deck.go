package entities

import (
	"time"
)

// FilteredDeck represents a filtered deck entity in the domain
// Filtered decks are dynamically generated based on search criteria
type FilteredDeck struct {
	id            int64
	userID        int64
	name          string
	searchFilter  string
	secondFilter  *string
	limitCards    int
	orderBy       string
	reschedule    bool
	createdAt     time.Time
	updatedAt     time.Time
	lastRebuildAt *time.Time
	deletedAt     *time.Time
}

// Getters
func (fd *FilteredDeck) GetID() int64 {
	return fd.id
}

func (fd *FilteredDeck) GetUserID() int64 {
	return fd.userID
}

func (fd *FilteredDeck) GetName() string {
	return fd.name
}

func (fd *FilteredDeck) GetSearchFilter() string {
	return fd.searchFilter
}

func (fd *FilteredDeck) GetSecondFilter() *string {
	return fd.secondFilter
}

func (fd *FilteredDeck) GetLimitCards() int {
	return fd.limitCards
}

func (fd *FilteredDeck) GetOrderBy() string {
	return fd.orderBy
}

func (fd *FilteredDeck) GetReschedule() bool {
	return fd.reschedule
}

func (fd *FilteredDeck) GetCreatedAt() time.Time {
	return fd.createdAt
}

func (fd *FilteredDeck) GetUpdatedAt() time.Time {
	return fd.updatedAt
}

func (fd *FilteredDeck) GetLastRebuildAt() *time.Time {
	return fd.lastRebuildAt
}

func (fd *FilteredDeck) GetDeletedAt() *time.Time {
	return fd.deletedAt
}

// Setters
func (fd *FilteredDeck) SetID(id int64) {
	fd.id = id
}

func (fd *FilteredDeck) SetUserID(userID int64) {
	fd.userID = userID
}

func (fd *FilteredDeck) SetName(name string) {
	fd.name = name
}

func (fd *FilteredDeck) SetSearchFilter(searchFilter string) {
	fd.searchFilter = searchFilter
}

func (fd *FilteredDeck) SetSecondFilter(secondFilter *string) {
	fd.secondFilter = secondFilter
}

func (fd *FilteredDeck) SetLimitCards(limitCards int) {
	fd.limitCards = limitCards
}

func (fd *FilteredDeck) SetOrderBy(orderBy string) {
	fd.orderBy = orderBy
}

func (fd *FilteredDeck) SetReschedule(reschedule bool) {
	fd.reschedule = reschedule
}

func (fd *FilteredDeck) SetCreatedAt(createdAt time.Time) {
	fd.createdAt = createdAt
}

func (fd *FilteredDeck) SetUpdatedAt(updatedAt time.Time) {
	fd.updatedAt = updatedAt
}

func (fd *FilteredDeck) SetLastRebuildAt(lastRebuildAt *time.Time) {
	fd.lastRebuildAt = lastRebuildAt
}

func (fd *FilteredDeck) SetDeletedAt(deletedAt *time.Time) {
	fd.deletedAt = deletedAt
}

// IsActive checks if the filtered deck is active (not deleted)
func (fd *FilteredDeck) IsActive() bool {
	return fd.deletedAt == nil
}

// NeedsRebuild checks if the filtered deck needs to be rebuilt
// This is a domain method - actual rebuild logic should be in service layer
func (fd *FilteredDeck) NeedsRebuild() bool {
	return fd.IsActive() && (fd.lastRebuildAt == nil || fd.updatedAt.After(*fd.lastRebuildAt))
}

