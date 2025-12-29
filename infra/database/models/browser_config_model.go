package models

import (
	"time"
)

// BrowserConfigModel represents the browser_config table structure in the database
type BrowserConfigModel struct {
	ID            int64
	UserID        int64 // Unique
	VisibleColumns string // TEXT[] stored as string (PostgreSQL array format)
	ColumnWidths   string // JSONB stored as string
	SortColumn     string
	SortDirection  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

