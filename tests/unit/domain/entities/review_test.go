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
			review: &entities.Review{
				Rating: 1,
			},
			expected: true,
		},
		{
			name: "valid rating 4",
			review: &entities.Review{
				Rating: 4,
			},
			expected: true,
		},
		{
			name: "invalid rating 0",
			review: &entities.Review{
				Rating: 0,
			},
			expected: false,
		},
		{
			name: "invalid rating 5",
			review: &entities.Review{
				Rating: 5,
			},
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
			review := &entities.Review{Rating: tt.rating}
			got := review.GetRatingName()
			if got != tt.expected {
				t.Errorf("Review.GetRatingName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReview_ReviewType(t *testing.T) {
	review := &entities.Review{
		Type:      valueobjects.ReviewTypeLearn,
		CreatedAt: time.Now(),
	}

	if review.Type != valueobjects.ReviewTypeLearn {
		t.Errorf("Review.Type = %v, want ReviewTypeLearn", review.Type)
	}
}

