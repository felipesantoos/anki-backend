package entities
import (
	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
)

import (
	"testing"
	"time"
)

func TestSharedDeck_IsActive(t *testing.T) {
	tests := []struct {
		name       string
		sharedDeck *shareddeck.SharedDeck
		expected   bool
	}{
		{
			name: "active shared deck",
			sharedDeck: func() *shareddeck.SharedDeck {
				sd := &shareddeck.SharedDeck{}
				sd.SetDeletedAt(nil)
				return sd
			}(),
			expected: true,
		},
		{
			name: "deleted shared deck",
			sharedDeck: func() *shareddeck.SharedDeck {
				sd := &shareddeck.SharedDeck{}
				sd.SetDeletedAt(timePtr(time.Now()))
				return sd
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sharedDeck.IsActive()
			if got != tt.expected {
				t.Errorf("SharedDeck.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSharedDeck_UpdateRating(t *testing.T) {
	sd := &shareddeck.SharedDeck{}
	sd.SetRatingAverage(0.0)
	sd.SetRatingCount(0)
	sd.SetUpdatedAt(time.Now())

	// First rating
	sd.UpdateRating(4.0)
	if sd.GetRatingAverage() != 4.0 {
		t.Errorf("SharedDeck.UpdateRating() average = %v, want 4.0", sd.GetRatingAverage())
	}
	if sd.GetRatingCount() != 1 {
		t.Errorf("SharedDeck.UpdateRating() count = %v, want 1", sd.GetRatingCount())
	}

	// Second rating
	sd.UpdateRating(5.0)
	expectedAvg := (4.0 + 5.0) / 2.0
	if sd.GetRatingAverage() != expectedAvg {
		t.Errorf("SharedDeck.UpdateRating() average = %v, want %v", sd.GetRatingAverage(), expectedAvg)
	}
	if sd.GetRatingCount() != 2 {
		t.Errorf("SharedDeck.UpdateRating() count = %v, want 2", sd.GetRatingCount())
	}

	// Invalid rating (should not update)
	originalAvg := sd.GetRatingAverage()
	originalCount := sd.GetRatingCount()
	sd.UpdateRating(6.0) // Invalid (> 5)
	if sd.GetRatingAverage() != originalAvg || sd.GetRatingCount() != originalCount {
		t.Errorf("SharedDeck.UpdateRating() should not accept invalid rating")
	}
}

func TestSharedDeck_IncrementDownloadCount(t *testing.T) {
	sd := &shareddeck.SharedDeck{}
	sd.SetDownloadCount(10)
	sd.SetUpdatedAt(time.Now())

	originalUpdatedAt := sd.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)

	sd.IncrementDownloadCount()
	if sd.GetDownloadCount() != 11 {
		t.Errorf("SharedDeck.IncrementDownloadCount() count = %v, want 11", sd.GetDownloadCount())
	}

	if sd.GetUpdatedAt().Equal(originalUpdatedAt) {
		t.Errorf("SharedDeck.IncrementDownloadCount() should update UpdatedAt")
	}
}


