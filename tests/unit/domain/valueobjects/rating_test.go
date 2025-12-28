package valueobjects

import (
	"errors"
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestNewRating(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		min     int
		max     int
		wantErr bool
	}{
		{
			name:    "valid rating in range",
			value:   3,
			min:     1,
			max:     5,
			wantErr: false,
		},
		{
			name:    "valid rating at min",
			value:   1,
			min:     1,
			max:     5,
			wantErr: false,
		},
		{
			name:    "valid rating at max",
			value:   5,
			min:     1,
			max:     5,
			wantErr: false,
		},
		{
			name:    "invalid rating below min",
			value:   0,
			min:     1,
			max:     5,
			wantErr: true,
		},
		{
			name:    "invalid rating above max",
			value:   6,
			min:     1,
			max:     5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, err := valueobjects.NewRating(tt.value, tt.min, tt.max)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRating() expected error, got nil")
					return
				}
				if !errors.Is(err, valueobjects.ErrRatingOutOfRange) {
					t.Errorf("NewRating() error = %v, want ErrRatingOutOfRange", err)
				}
				return
			}

			if err != nil {
				t.Errorf("NewRating() unexpected error = %v", err)
				return
			}

			if rating.Value() != tt.value {
				t.Errorf("NewRating() value = %v, want %v", rating.Value(), tt.value)
			}
		})
	}
}

func TestNewReviewRating(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{
			name:    "valid review rating",
			value:   3,
			wantErr: false,
		},
		{
			name:    "valid rating 1",
			value:   1,
			wantErr: false,
		},
		{
			name:    "valid rating 4",
			value:   4,
			wantErr: false,
		},
		{
			name:    "invalid rating 0",
			value:   0,
			wantErr: true,
		},
		{
			name:    "invalid rating 5",
			value:   5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, err := valueobjects.NewReviewRating(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewReviewRating() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewReviewRating() unexpected error = %v", err)
				return
			}

			if rating.Value() != tt.value {
				t.Errorf("NewReviewRating() value = %v, want %v", rating.Value(), tt.value)
			}
		})
	}
}

func TestNewSharedDeckRating(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{
			name:    "valid shared deck rating",
			value:   3,
			wantErr: false,
		},
		{
			name:    "valid rating 1",
			value:   1,
			wantErr: false,
		},
		{
			name:    "valid rating 5",
			value:   5,
			wantErr: false,
		},
		{
			name:    "invalid rating 0",
			value:   0,
			wantErr: true,
		},
		{
			name:    "invalid rating 6",
			value:   6,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, err := valueobjects.NewSharedDeckRating(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewSharedDeckRating() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewSharedDeckRating() unexpected error = %v", err)
				return
			}

			if rating.Value() != tt.value {
				t.Errorf("NewSharedDeckRating() value = %v, want %v", rating.Value(), tt.value)
			}
		})
	}
}

func TestRating_IsValid(t *testing.T) {
	rating, err := valueobjects.NewRating(3, 1, 5)
	if err != nil {
		t.Fatalf("NewRating() error = %v", err)
	}

	if !rating.IsValid() {
		t.Errorf("Rating.IsValid() = false, want true")
	}
}

func TestRating_String(t *testing.T) {
	rating, err := valueobjects.NewRating(3, 1, 5)
	if err != nil {
		t.Fatalf("NewRating() error = %v", err)
	}

	if rating.String() != "3" {
		t.Errorf("Rating.String() = %v, want '3'", rating.String())
	}
}

func TestRating_Equals(t *testing.T) {
	rating1, err := valueobjects.NewRating(3, 1, 5)
	if err != nil {
		t.Fatalf("NewRating() error = %v", err)
	}

	rating2, err := valueobjects.NewRating(3, 1, 5)
	if err != nil {
		t.Fatalf("NewRating() error = %v", err)
	}

	rating3, err := valueobjects.NewRating(4, 1, 5)
	if err != nil {
		t.Fatalf("NewRating() error = %v", err)
	}

	if !rating1.Equals(rating2) {
		t.Errorf("Rating.Equals() = false, want true for same ratings")
	}

	if rating1.Equals(rating3) {
		t.Errorf("Rating.Equals() = true, want false for different ratings")
	}
}

