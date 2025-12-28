package entities

import (
	"time"
)

// BrowserConfig represents browser configuration entity in the domain
// It stores user preferences for the card browser interface
type BrowserConfig struct {
	id             int64
	userID         int64 // Unique
	visibleColumns []string
	columnWidths   string // JSONB in database
	sortColumn     *string
	sortDirection  string // asc, desc
	createdAt      time.Time
	updatedAt      time.Time
}

// Getters
func (bc *BrowserConfig) GetID() int64 {
	return bc.id
}

func (bc *BrowserConfig) GetUserID() int64 {
	return bc.userID
}

func (bc *BrowserConfig) GetVisibleColumns() []string {
	return bc.visibleColumns
}

func (bc *BrowserConfig) GetColumnWidths() string {
	return bc.columnWidths
}

func (bc *BrowserConfig) GetSortColumn() *string {
	return bc.sortColumn
}

func (bc *BrowserConfig) GetSortDirection() string {
	return bc.sortDirection
}

func (bc *BrowserConfig) GetCreatedAt() time.Time {
	return bc.createdAt
}

func (bc *BrowserConfig) GetUpdatedAt() time.Time {
	return bc.updatedAt
}

// Setters
func (bc *BrowserConfig) SetID(id int64) {
	bc.id = id
}

func (bc *BrowserConfig) SetUserID(userID int64) {
	bc.userID = userID
}

func (bc *BrowserConfig) SetVisibleColumns(visibleColumns []string) {
	bc.visibleColumns = visibleColumns
}

func (bc *BrowserConfig) SetColumnWidths(columnWidths string) {
	bc.columnWidths = columnWidths
}

func (bc *BrowserConfig) SetSortColumn(sortColumn *string) {
	bc.sortColumn = sortColumn
}

func (bc *BrowserConfig) SetSortDirection(sortDirection string) {
	bc.sortDirection = sortDirection
}

func (bc *BrowserConfig) SetCreatedAt(createdAt time.Time) {
	bc.createdAt = createdAt
}

func (bc *BrowserConfig) SetUpdatedAt(updatedAt time.Time) {
	bc.updatedAt = updatedAt
}

