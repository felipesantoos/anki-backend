package models

import (
	"time"
)

// ReviewModel represents the reviews table structure in the database
type ReviewModel struct {
	ID        int64
	CardID    int64
	Rating    int
	Interval  int
	Ease      int
	TimeMs    int
	Type      string // review_type enum stored as string
	CreatedAt time.Time
}

