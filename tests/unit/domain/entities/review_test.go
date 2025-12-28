package entities
import (
	"github.com/felipesantos/anki-backend/core/domain/entities"
)

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestReview_IsValidRating(t *testing.T) {
	tests := []struct {
		name     string
		review   *entities.Review
		expected bool
	}{
		{
			name: "valid rating 1",
			review: func() *entities.Review {
				r := &entities.Review{}
				r.SetRating(1)
				return r
			}(),
			expected: true,
		},
		{
			name: "valid rating 4",
			review: func() *entities.Review {
				r := &entities.Review{}
				r.SetRating(4)
				return r
			}(),
			expected: true,
		},
		{
			name: "invalid rating 0",
			review: func() *entities.Review {
				r := &entities.Review{}
				r.SetRating(0)
				return r
			}(),
			expected: false,
		},
		{
			name: "invalid rating 5",
			review: func() *entities.Review {
				r := &entities.Review{}
				r.SetRating(5)
				return r
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.review.IsValidRating()
			if got != tt.expected {
				t.Errorf("Review.IsValidRating() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReview_GetRatingName(t *testing.T) {
	tests := []struct {
		name     string
		rating   int
		expected string
	}{
		{
			name:     "rating 1 - Again",
			rating:   1,
			expected: "Again",
		},
		{
			name:     "rating 2 - Hard",
			rating:   2,
			expected: "Hard",
		},
		{
			name:     "rating 3 - Good",
			rating:   3,
			expected: "Good",
		},
		{
			name:     "rating 4 - Easy",
			rating:   4,
			expected: "Easy",
		},
		{
			name:     "invalid rating",
			rating:   0,
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := &entities.Review{}
			review.SetRating(tt.rating)
			got := review.GetRatingName()
			if got != tt.expected {
				t.Errorf("Review.GetRatingName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReview_ReviewType(t *testing.T) {
	review := &entities.Review{}
	review.SetType(valueobjects.ReviewTypeLearn)
	review.SetCreatedAt(time.Now())

	if review.GetType() != valueobjects.ReviewTypeLearn {
		t.Errorf("Review.GetType() = %v, want ReviewTypeLearn", review.GetType())
	}
}

