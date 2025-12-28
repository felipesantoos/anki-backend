package entities

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestCard_IsDue(t *testing.T) {
	now := int64(1000000000000) // Timestamp in milliseconds

	tests := []struct {
		name     string
		card     *entities.Card
		now      int64
		expected bool
	}{
		{
			name: "new card is always due",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateNew)
				c.SetSuspended(false)
				c.SetBuried(false)
				return c
			}(),
			now:      now,
			expected: true,
		},
		{
			name: "review card due",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateReview)
				c.SetDue(now - 1000)
				c.SetSuspended(false)
				c.SetBuried(false)
				return c
			}(),
			now:      now,
			expected: true,
		},
		{
			name: "review card not due",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateReview)
				c.SetDue(now + 1000)
				c.SetSuspended(false)
				c.SetBuried(false)
				return c
			}(),
			now:      now,
			expected: false,
		},
		{
			name: "suspended card not due",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateReview)
				c.SetDue(now - 1000)
				c.SetSuspended(true)
				c.SetBuried(false)
				return c
			}(),
			now:      now,
			expected: false,
		},
		{
			name: "buried card not due",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateReview)
				c.SetDue(now - 1000)
				c.SetSuspended(false)
				c.SetBuried(true)
				return c
			}(),
			now:      now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.card.IsDue(tt.now)
			if got != tt.expected {
				t.Errorf("Card.IsDue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCard_IsStudyable(t *testing.T) {
	tests := []struct {
		name     string
		card     *entities.Card
		expected bool
	}{
		{
			name: "studyable card",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetSuspended(false)
				c.SetBuried(false)
				return c
			}(),
			expected: true,
		},
		{
			name: "suspended card not studyable",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetSuspended(true)
				c.SetBuried(false)
				return c
			}(),
			expected: false,
		},
		{
			name: "buried card not studyable",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetSuspended(false)
				c.SetBuried(true)
				return c
			}(),
			expected: false,
		},
		{
			name: "suspended and buried card not studyable",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetSuspended(true)
				c.SetBuried(true)
				return c
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.card.IsStudyable()
			if got != tt.expected {
				t.Errorf("Card.IsStudyable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCard_StateChecks(t *testing.T) {
	tests := []struct {
		name     string
		card     *entities.Card
		isNew    bool
		isLearn  bool
		isReview bool
		isRelearn bool
	}{
		{
			name: "new card",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateNew)
				return c
			}(),
			isNew:     true,
			isLearn:   false,
			isReview:  false,
			isRelearn: false,
		},
		{
			name: "learning card",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateLearn)
				return c
			}(),
			isNew:     false,
			isLearn:   true,
			isReview:  false,
			isRelearn: false,
		},
		{
			name: "review card",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateReview)
				return c
			}(),
			isNew:     false,
			isLearn:   false,
			isReview:  true,
			isRelearn: false,
		},
		{
			name: "relearning card",
			card: func() *entities.Card {
				c := &entities.Card{}
				c.SetState(valueobjects.CardStateRelearn)
				return c
			}(),
			isNew:     false,
			isLearn:   false,
			isReview:  false,
			isRelearn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.card.IsNew() != tt.isNew {
				t.Errorf("Card.IsNew() = %v, want %v", tt.card.IsNew(), tt.isNew)
			}
			if tt.card.IsLearning() != tt.isLearn {
				t.Errorf("Card.IsLearning() = %v, want %v", tt.card.IsLearning(), tt.isLearn)
			}
			if tt.card.IsReview() != tt.isReview {
				t.Errorf("Card.IsReview() = %v, want %v", tt.card.IsReview(), tt.isReview)
			}
			if tt.card.IsRelearning() != tt.isRelearn {
				t.Errorf("Card.IsRelearning() = %v, want %v", tt.card.IsRelearning(), tt.isRelearn)
			}
		})
	}
}

func TestCard_GetNextReviewTime(t *testing.T) {
	due := int64(1000000000000)
	card := &entities.Card{}
	card.SetDue(due)

	if card.GetNextReviewTime() != due {
		t.Errorf("Card.GetNextReviewTime() = %v, want %v", card.GetNextReviewTime(), due)
	}
}

func TestCard_Suspend(t *testing.T) {
	card := &entities.Card{}
	card.SetSuspended(false)
	card.SetUpdatedAt(time.Now())

	card.Suspend()
	if !card.GetSuspended() {
		t.Errorf("Card.Suspend() failed to suspend card")
	}

	// Suspend again (should be idempotent)
	card.Suspend()
	if !card.GetSuspended() {
		t.Errorf("Card.Suspend() failed to keep card suspended")
	}
}

func TestCard_Unsuspend(t *testing.T) {
	card := &entities.Card{}
	card.SetSuspended(true)
	card.SetUpdatedAt(time.Now())

	card.Unsuspend()
	if card.GetSuspended() {
		t.Errorf("Card.Unsuspend() failed to unsuspend card")
	}

	// Unsuspend again (should be idempotent)
	card.Unsuspend()
	if card.GetSuspended() {
		t.Errorf("Card.Unsuspend() failed to keep card unsuspended")
	}
}

func TestCard_Bury(t *testing.T) {
	card := &entities.Card{}
	card.SetBuried(false)
	card.SetUpdatedAt(time.Now())

	card.Bury()
	if !card.GetBuried() {
		t.Errorf("Card.Bury() failed to bury card")
	}

	// Bury again (should be idempotent)
	card.Bury()
	if !card.GetBuried() {
		t.Errorf("Card.Bury() failed to keep card buried")
	}
}

func TestCard_Unbury(t *testing.T) {
	card := &entities.Card{}
	card.SetBuried(true)
	card.SetUpdatedAt(time.Now())

	card.Unbury()
	if card.GetBuried() {
		t.Errorf("Card.Unbury() failed to unbury card")
	}

	// Unbury again (should be idempotent)
	card.Unbury()
	if card.GetBuried() {
		t.Errorf("Card.Unbury() failed to keep card unburied")
	}
}

func TestCard_SetFlag(t *testing.T) {
	card := &entities.Card{}
	card.SetFlag(0)
	card.SetUpdatedAt(time.Now())

	tests := []struct {
		name    string
		flag    int
		wantErr bool
	}{
		{
			name:    "valid flag 0",
			flag:    0,
			wantErr: false,
		},
		{
			name:    "valid flag 7",
			flag:    7,
			wantErr: false,
		},
		{
			name:    "invalid flag -1",
			flag:    -1,
			wantErr: true,
		},
		{
			name:    "invalid flag 8",
			flag:    8,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := card.SetFlag(tt.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("Card.SetFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && card.GetFlag() != tt.flag {
				t.Errorf("Card.SetFlag() flag = %v, want %v", card.GetFlag(), tt.flag)
			}
		})
	}
}

