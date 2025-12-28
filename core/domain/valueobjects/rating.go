package valueobjects

import (
	"errors"
	"fmt"
)

var (
	// ErrRatingOutOfRange is returned when rating is outside valid range
	ErrRatingOutOfRange = errors.New("rating is out of valid range")
)

// Rating represents a rating value (1-5 for shared decks, 1-4 for reviews)
type Rating struct {
	value int
	min   int
	max   int
}

// NewRating creates a new Rating value object
// min and max define the valid range (typically 1-5 or 1-4)
func NewRating(value, min, max int) (Rating, error) {
	if value < min || value > max {
		return Rating{}, fmt.Errorf("%w: %d (valid range: %d-%d)", ErrRatingOutOfRange, value, min, max)
	}

	return Rating{
		value: value,
		min:   min,
		max:   max,
	}, nil
}

// NewReviewRating creates a rating for reviews (1-4)
func NewReviewRating(value int) (Rating, error) {
	return NewRating(value, 1, 4)
}

// NewSharedDeckRating creates a rating for shared decks (1-5)
func NewSharedDeckRating(value int) (Rating, error) {
	return NewRating(value, 1, 5)
}

// Value returns the rating value as an integer
func (r Rating) Value() int {
	return r.value
}

// IsValid checks if the rating is within its valid range
func (r Rating) IsValid() bool {
	return r.value >= r.min && r.value <= r.max
}

// String returns the string representation of the rating
func (r Rating) String() string {
	return fmt.Sprintf("%d", r.value)
}

// Equals compares two Rating value objects for equality
func (r Rating) Equals(other Rating) bool {
	return r.value == other.value && r.min == other.min && r.max == other.max
}

