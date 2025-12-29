package card

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
	id            int64
	noteID        int64
	cardTypeID    int
	deckID        int64
	homeDeckID    *int64
	due           int64 // Timestamp in milliseconds
	interval      int   // Days (or negative seconds for learning)
	ease          int   // Permille (2500 = 2.5x)
	lapses        int
	reps          int
	state         valueobjects.CardState
	position      int
	flag          int // 0-7
	suspended     bool
	buried        bool
	stability     *float64 // FSRS stability (in days)
	difficulty    *float64 // FSRS difficulty (0.0-1.0)
	lastReviewAt  *time.Time
	createdAt     time.Time
	updatedAt     time.Time
}

// Getters
func (c *Card) GetID() int64 {
	return c.id
}

func (c *Card) GetNoteID() int64 {
	return c.noteID
}

func (c *Card) GetCardTypeID() int {
	return c.cardTypeID
}

func (c *Card) GetDeckID() int64 {
	return c.deckID
}

func (c *Card) GetHomeDeckID() *int64 {
	return c.homeDeckID
}

func (c *Card) GetDue() int64 {
	return c.due
}

func (c *Card) GetInterval() int {
	return c.interval
}

func (c *Card) GetEase() int {
	return c.ease
}

func (c *Card) GetLapses() int {
	return c.lapses
}

func (c *Card) GetReps() int {
	return c.reps
}

func (c *Card) GetState() valueobjects.CardState {
	return c.state
}

func (c *Card) GetPosition() int {
	return c.position
}

func (c *Card) GetFlag() int {
	return c.flag
}

func (c *Card) GetSuspended() bool {
	return c.suspended
}

func (c *Card) GetBuried() bool {
	return c.buried
}

func (c *Card) GetStability() *float64 {
	return c.stability
}

func (c *Card) GetDifficulty() *float64 {
	return c.difficulty
}

func (c *Card) GetLastReviewAt() *time.Time {
	return c.lastReviewAt
}

func (c *Card) GetCreatedAt() time.Time {
	return c.createdAt
}

func (c *Card) GetUpdatedAt() time.Time {
	return c.updatedAt
}

// Setters
func (c *Card) SetID(id int64) {
	c.id = id
}

func (c *Card) SetNoteID(noteID int64) {
	c.noteID = noteID
}

func (c *Card) SetCardTypeID(cardTypeID int) {
	c.cardTypeID = cardTypeID
}

func (c *Card) SetDeckID(deckID int64) {
	c.deckID = deckID
}

func (c *Card) SetHomeDeckID(homeDeckID *int64) {
	c.homeDeckID = homeDeckID
}

func (c *Card) SetDue(due int64) {
	c.due = due
}

func (c *Card) SetInterval(interval int) {
	c.interval = interval
}

func (c *Card) SetEase(ease int) {
	c.ease = ease
}

func (c *Card) SetLapses(lapses int) {
	c.lapses = lapses
}

func (c *Card) SetReps(reps int) {
	c.reps = reps
}

func (c *Card) SetState(state valueobjects.CardState) {
	c.state = state
}

func (c *Card) SetPosition(position int) {
	c.position = position
}

func (c *Card) SetSuspended(suspended bool) {
	c.suspended = suspended
}

func (c *Card) SetBuried(buried bool) {
	c.buried = buried
}

func (c *Card) SetStability(stability *float64) {
	c.stability = stability
}

func (c *Card) SetDifficulty(difficulty *float64) {
	c.difficulty = difficulty
}

func (c *Card) SetLastReviewAt(lastReviewAt *time.Time) {
	c.lastReviewAt = lastReviewAt
}

func (c *Card) SetCreatedAt(createdAt time.Time) {
	c.createdAt = createdAt
}

func (c *Card) SetUpdatedAt(updatedAt time.Time) {
	c.updatedAt = updatedAt
}

// IsDue checks if the card is due for review at the given timestamp
func (c *Card) IsDue(now int64) bool {
	if c.suspended || c.buried {
		return false
	}

	switch c.state {
	case valueobjects.CardStateNew:
		return true
	case valueobjects.CardStateLearn, valueobjects.CardStateRelearn, valueobjects.CardStateReview:
		return c.due <= now
	default:
		return false
	}
}

// IsStudyable checks if the card can be studied (not suspended or buried)
func (c *Card) IsStudyable() bool {
	return !c.suspended && !c.buried
}

// IsNew checks if the card is in new state
func (c *Card) IsNew() bool {
	return c.state == valueobjects.CardStateNew
}

// IsLearning checks if the card is in learning state
func (c *Card) IsLearning() bool {
	return c.state == valueobjects.CardStateLearn
}

// IsReview checks if the card is in review state
func (c *Card) IsReview() bool {
	return c.state == valueobjects.CardStateReview
}

// IsRelearning checks if the card is in relearning state
func (c *Card) IsRelearning() bool {
	return c.state == valueobjects.CardStateRelearn
}

// GetNextReviewTime returns the next review time (due timestamp)
func (c *Card) GetNextReviewTime() int64 {
	return c.due
}

// Suspend suspends the card
func (c *Card) Suspend() {
	if !c.suspended {
		c.suspended = true
		c.updatedAt = time.Now()
	}
}

// Unsuspend removes suspension from the card
func (c *Card) Unsuspend() {
	if c.suspended {
		c.suspended = false
		c.updatedAt = time.Now()
	}
}

// Bury buries the card
func (c *Card) Bury() {
	if !c.buried {
		c.buried = true
		c.updatedAt = time.Now()
	}
}

// Unbury removes burial from the card
func (c *Card) Unbury() {
	if c.buried {
		c.buried = false
		c.updatedAt = time.Now()
	}
}

// SetFlag sets the flag value (0-7)
func (c *Card) SetFlag(flag int) error {
	if flag < 0 || flag > 7 {
		return ErrInvalidFlag
	}
	c.flag = flag
	c.updatedAt = time.Now()
	return nil
}

