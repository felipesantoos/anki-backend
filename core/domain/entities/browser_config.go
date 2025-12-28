package entities

import (
	"time"
)

// BrowserConfig represents browser configuration entity in the domain
// It stores user preferences for the card browser interface
type BrowserConfig struct {
	ID             int64
	UserID         int64 // Unique
	VisibleColumns []string
	ColumnWidths   string // JSONB in database
	SortColumn     *string
	SortDirection  string // asc, desc
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

