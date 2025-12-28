package valueobjects

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestCardState_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		state valueobjects.CardState
		want  bool
	}{
		{
			name:  "valid new",
			state: valueobjects.CardStateNew,
			want:  true,
		},
		{
			name:  "valid learn",
			state: valueobjects.CardStateLearn,
			want:  true,
		},
		{
			name:  "valid review",
			state: valueobjects.CardStateReview,
			want:  true,
		},
		{
			name:  "valid relearn",
			state: valueobjects.CardStateRelearn,
			want:  true,
		},
		{
			name:  "invalid state",
			state: valueobjects.CardState("invalid"),
			want:  false,
		},
		{
			name:  "empty state",
			state: valueobjects.CardState(""),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.IsValid()
			if got != tt.want {
				t.Errorf("CardState.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardState_String(t *testing.T) {
	if valueobjects.CardStateNew.String() != "new" {
		t.Errorf("CardStateNew.String() = %v, want 'new'", valueobjects.CardStateNew.String())
	}
	if valueobjects.CardStateLearn.String() != "learn" {
		t.Errorf("CardStateLearn.String() = %v, want 'learn'", valueobjects.CardStateLearn.String())
	}
	if valueobjects.CardStateReview.String() != "review" {
		t.Errorf("CardStateReview.String() = %v, want 'review'", valueobjects.CardStateReview.String())
	}
	if valueobjects.CardStateRelearn.String() != "relearn" {
		t.Errorf("CardStateRelearn.String() = %v, want 'relearn'", valueobjects.CardStateRelearn.String())
	}
}

func TestCardState_CanBeStudied(t *testing.T) {
	tests := []struct {
		name  string
		state valueobjects.CardState
		want  bool
	}{
		{
			name:  "new can be studied",
			state: valueobjects.CardStateNew,
			want:  true,
		},
		{
			name:  "learn can be studied",
			state: valueobjects.CardStateLearn,
			want:  true,
		},
		{
			name:  "review can be studied",
			state: valueobjects.CardStateReview,
			want:  true,
		},
		{
			name:  "relearn can be studied",
			state: valueobjects.CardStateRelearn,
			want:  true,
		},
		{
			name:  "invalid cannot be studied",
			state: valueobjects.CardState("invalid"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.CanBeStudied()
			if got != tt.want {
				t.Errorf("CardState.CanBeStudied() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardState_IsNew(t *testing.T) {
	if !valueobjects.CardStateNew.IsNew() {
		t.Errorf("CardStateNew.IsNew() = false, want true")
	}
	if valueobjects.CardStateLearn.IsNew() {
		t.Errorf("CardStateLearn.IsNew() = true, want false")
	}
}

func TestCardState_IsLearning(t *testing.T) {
	if !valueobjects.CardStateLearn.IsLearning() {
		t.Errorf("CardStateLearn.IsLearning() = false, want true")
	}
	if valueobjects.CardStateNew.IsLearning() {
		t.Errorf("CardStateNew.IsLearning() = true, want false")
	}
}

func TestCardState_IsReview(t *testing.T) {
	if !valueobjects.CardStateReview.IsReview() {
		t.Errorf("CardStateReview.IsReview() = false, want true")
	}
	if valueobjects.CardStateNew.IsReview() {
		t.Errorf("CardStateNew.IsReview() = true, want false")
	}
}

func TestCardState_IsRelearning(t *testing.T) {
	if !valueobjects.CardStateRelearn.IsRelearning() {
		t.Errorf("CardStateRelearn.IsRelearning() = false, want true")
	}
	if valueobjects.CardStateNew.IsRelearning() {
		t.Errorf("CardStateNew.IsRelearning() = true, want false")
	}
}

