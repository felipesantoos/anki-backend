package entities

import (
	"time"
)

// SharedDeck represents a shared deck entity in the domain
// It stores information about publicly shared decks
type SharedDeck struct {
	ID             int64
	AuthorID       int64
	Name           string
	Description    *string
	Category       *string
	PackagePath    string
	PackageSize    int64
	DownloadCount  int
	RatingAverage  float64
	RatingCount    int
	Tags           []string
	IsFeatured     bool
	IsPublic       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// IsActive checks if the shared deck is active (not deleted)
func (sd *SharedDeck) IsActive() bool {
	return sd.DeletedAt == nil
}

// UpdateRating updates the average rating when a new rating is added
func (sd *SharedDeck) UpdateRating(newRating float64) {
	if newRating < 1 || newRating > 5 {
		return
	}

	// Calculate new average: (old_average * old_count + new_rating) / (old_count + 1)
	totalRating := sd.RatingAverage*float64(sd.RatingCount) + newRating
	sd.RatingCount++
	sd.RatingAverage = totalRating / float64(sd.RatingCount)
	sd.UpdatedAt = time.Now()
}

// IncrementDownloadCount increments the download count
func (sd *SharedDeck) IncrementDownloadCount() {
	sd.DownloadCount++
	sd.UpdatedAt = time.Now()
}

