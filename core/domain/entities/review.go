package entities

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// Review represents a review entity in the domain
// A review records a study session for a card
type Review struct {
	ID        int64
	CardID    int64
	Rating    int // 1-4: Again, Hard, Good, Easy
	Interval  int // New interval after review (days or negative seconds)
	Ease      int // New ease factor after review (permille)
	TimeMs    int // Time spent on review (milliseconds)
	Type      valueobjects.ReviewType
	CreatedAt time.Time
}

// IsValidRating checks if the rating is valid (1-4)
func (r *Review) IsValidRating() bool {
	return r.Rating >= 1 && r.Rating <= 4
}

// GetRatingName returns the name of the rating
func (r *Review) GetRatingName() string {
	switch r.Rating {
	case 1:
		return "Again"
	case 2:
		return "Hard"
	case 3:
		return "Good"
	case 4:
		return "Easy"
	default:
		return "Unknown"
	}
}

