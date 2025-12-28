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
			sharedDeck: &entities.SharedDeck{
				DeletedAt: nil,
			},
			expected: true,
		},
		{
			name: "deleted shared deck",
			sharedDeck: &entities.SharedDeck{
				DeletedAt: timePtr(time.Now()),
			},
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
	deck := &entities.SharedDeck{
		RatingAverage: 0.0,
		RatingCount:   0,
		UpdatedAt:     time.Now(),
	}

	// First rating
	deck.UpdateRating(4.0)
	if deck.RatingAverage != 4.0 {
		t.Errorf("SharedDeck.UpdateRating() average = %v, want 4.0", deck.RatingAverage)
	}
	if deck.RatingCount != 1 {
		t.Errorf("SharedDeck.UpdateRating() count = %v, want 1", deck.RatingCount)
	}

	// Second rating
	deck.UpdateRating(5.0)
	expectedAvg := (4.0 + 5.0) / 2.0
	if deck.RatingAverage != expectedAvg {
		t.Errorf("SharedDeck.UpdateRating() average = %v, want %v", deck.RatingAverage, expectedAvg)
	}
	if deck.RatingCount != 2 {
		t.Errorf("SharedDeck.UpdateRating() count = %v, want 2", deck.RatingCount)
	}

	// Invalid rating (should not update)
	originalAvg := deck.RatingAverage
	originalCount := deck.RatingCount
	deck.UpdateRating(6.0) // Invalid (> 5)
	if deck.RatingAverage != originalAvg || deck.RatingCount != originalCount {
		t.Errorf("SharedDeck.UpdateRating() should not accept invalid rating")
	}
}

func TestSharedDeck_IncrementDownloadCount(t *testing.T) {
	deck := &entities.SharedDeck{
		DownloadCount: 10,
		UpdatedAt:    time.Now(),
	}

	originalUpdatedAt := deck.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	deck.IncrementDownloadCount()
	if deck.DownloadCount != 11 {
		t.Errorf("SharedDeck.IncrementDownloadCount() count = %v, want 11", deck.DownloadCount)
	}

	if deck.UpdatedAt.Equal(originalUpdatedAt) {
		t.Errorf("SharedDeck.IncrementDownloadCount() should update UpdatedAt")
	}
}


