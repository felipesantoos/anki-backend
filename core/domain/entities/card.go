package entities

import (
	"errors"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	// ErrInvalidFlag is returned when flag value is outside valid range (0-7)
	ErrInvalidFlag = errors.New("flag must be between 0 and 7")
)

// Card represents a card entity in the domain
// A card is generated from a note and represents a single question/answer pair
type Card struct {
	ID            int64
	NoteID        int64
	CardTypeID    int
	DeckID        int64
	HomeDeckID    *int64
	Due           int64 // Timestamp in milliseconds
	Interval      int   // Days (or negative seconds for learning)
	Ease          int   // Permille (2500 = 2.5x)
	Lapses        int
	Reps          int
	State         valueobjects.CardState
	Position      int
	Flag          int // 0-7
	Suspended     bool
	Buried        bool
	Stability     *float64 // FSRS stability (in days)
	Difficulty    *float64 // FSRS difficulty (0.0-1.0)
	LastReviewAt  *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsDue checks if the card is due for review at the given timestamp
func (c *Card) IsDue(now int64) bool {
	if c.Suspended || c.Buried {
		return false
	}

	switch c.State {
	case valueobjects.CardStateNew:
		return true
	case valueobjects.CardStateLearn, valueobjects.CardStateRelearn, valueobjects.CardStateReview:
		return c.Due <= now
	default:
		return false
	}
}

// IsStudyable checks if the card can be studied (not suspended or buried)
func (c *Card) IsStudyable() bool {
	return !c.Suspended && !c.Buried
}

// IsNew checks if the card is in new state
func (c *Card) IsNew() bool {
	return c.State == valueobjects.CardStateNew
}

// IsLearning checks if the card is in learning state
func (c *Card) IsLearning() bool {
	return c.State == valueobjects.CardStateLearn
}

// IsReview checks if the card is in review state
func (c *Card) IsReview() bool {
	return c.State == valueobjects.CardStateReview
}

// IsRelearning checks if the card is in relearning state
func (c *Card) IsRelearning() bool {
	return c.State == valueobjects.CardStateRelearn
}

// GetNextReviewTime returns the next review time (due timestamp)
func (c *Card) GetNextReviewTime() int64 {
	return c.Due
}

// Suspend suspends the card
func (c *Card) Suspend() {
	if !c.Suspended {
		c.Suspended = true
		c.UpdatedAt = time.Now()
	}
}

// Unsuspend removes suspension from the card
func (c *Card) Unsuspend() {
	if c.Suspended {
		c.Suspended = false
		c.UpdatedAt = time.Now()
	}
}

// Bury buries the card
func (c *Card) Bury() {
	if !c.Buried {
		c.Buried = true
		c.UpdatedAt = time.Now()
	}
}

// Unbury removes burial from the card
func (c *Card) Unbury() {
	if c.Buried {
		c.Buried = false
		c.UpdatedAt = time.Now()
	}
}

// SetFlag sets the flag value (0-7)
func (c *Card) SetFlag(flag int) error {
	if flag < 0 || flag > 7 {
		return ErrInvalidFlag
	}
	c.Flag = flag
	c.UpdatedAt = time.Now()
	return nil
}

