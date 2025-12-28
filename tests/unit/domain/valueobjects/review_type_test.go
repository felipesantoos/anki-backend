package valueobjects

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestReviewType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		reviewType valueobjects.ReviewType
		want       bool
	}{
		{
			name:       "valid learn",
			reviewType: valueobjects.ReviewTypeLearn,
			want:       true,
		},
		{
			name:       "valid review",
			reviewType: valueobjects.ReviewTypeReview,
			want:       true,
		},
		{
			name:       "valid relearn",
			reviewType: valueobjects.ReviewTypeRelearn,
			want:       true,
		},
		{
			name:       "valid cram",
			reviewType: valueobjects.ReviewTypeCram,
			want:       true,
		},
		{
			name:       "invalid type",
			reviewType: valueobjects.ReviewType("invalid"),
			want:       false,
		},
		{
			name:       "empty type",
			reviewType: valueobjects.ReviewType(""),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.reviewType.IsValid()
			if got != tt.want {
				t.Errorf("ReviewType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReviewType_String(t *testing.T) {
	if valueobjects.ReviewTypeLearn.String() != "learn" {
		t.Errorf("ReviewTypeLearn.String() = %v, want 'learn'", valueobjects.ReviewTypeLearn.String())
	}
	if valueobjects.ReviewTypeReview.String() != "review" {
		t.Errorf("ReviewTypeReview.String() = %v, want 'review'", valueobjects.ReviewTypeReview.String())
	}
	if valueobjects.ReviewTypeRelearn.String() != "relearn" {
		t.Errorf("ReviewTypeRelearn.String() = %v, want 'relearn'", valueobjects.ReviewTypeRelearn.String())
	}
	if valueobjects.ReviewTypeCram.String() != "cram" {
		t.Errorf("ReviewTypeCram.String() = %v, want 'cram'", valueobjects.ReviewTypeCram.String())
	}
}

