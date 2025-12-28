package entities
import (
	"github.com/felipesantos/anki-backend/core/domain/entities"
)

import (
	"testing"
	"time"
)

func TestSharedDeck_IsActive(t *testing.T) {
	tests := []struct {
		name       string
		sharedDeck *entities.SharedDeck
		expected   bool
	}{
		{
			name: "active shared deck",
			sharedDeck: func() *entities.SharedDeck {
				sd := &entities.SharedDeck{}
				sd.SetDeletedAt(nil)
				return sd
			}(),
			expected: true,
		},
		{
			name: "deleted shared deck",
			sharedDeck: func() *entities.SharedDeck {
				sd := &entities.SharedDeck{}
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
	deck := &entities.SharedDeck{}
	deck.SetRatingAverage(0.0)
	deck.SetRatingCount(0)
	deck.SetUpdatedAt(time.Now())

	// First rating
	deck.UpdateRating(4.0)
	if deck.GetRatingAverage() != 4.0 {
		t.Errorf("SharedDeck.UpdateRating() average = %v, want 4.0", deck.GetRatingAverage())
	}
	if deck.GetRatingCount() != 1 {
		t.Errorf("SharedDeck.UpdateRating() count = %v, want 1", deck.GetRatingCount())
	}

	// Second rating
	deck.UpdateRating(5.0)
	expectedAvg := (4.0 + 5.0) / 2.0
	if deck.GetRatingAverage() != expectedAvg {
		t.Errorf("SharedDeck.UpdateRating() average = %v, want %v", deck.GetRatingAverage(), expectedAvg)
	}
	if deck.GetRatingCount() != 2 {
		t.Errorf("SharedDeck.UpdateRating() count = %v, want 2", deck.GetRatingCount())
	}

	// Invalid rating (should not update)
	originalAvg := deck.GetRatingAverage()
	originalCount := deck.GetRatingCount()
	deck.UpdateRating(6.0) // Invalid (> 5)
	if deck.GetRatingAverage() != originalAvg || deck.GetRatingCount() != originalCount {
		t.Errorf("SharedDeck.UpdateRating() should not accept invalid rating")
	}
}

func TestSharedDeck_IncrementDownloadCount(t *testing.T) {
	deck := &entities.SharedDeck{}
	deck.SetDownloadCount(10)
	deck.SetUpdatedAt(time.Now())

	originalUpdatedAt := deck.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)

	deck.IncrementDownloadCount()
	if deck.GetDownloadCount() != 11 {
		t.Errorf("SharedDeck.IncrementDownloadCount() count = %v, want 11", deck.GetDownloadCount())
	}

	if deck.GetUpdatedAt().Equal(originalUpdatedAt) {
		t.Errorf("SharedDeck.IncrementDownloadCount() should update UpdatedAt")
	}
}


