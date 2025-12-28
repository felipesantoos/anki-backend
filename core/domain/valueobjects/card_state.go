package valueobjects

// CardState represents the state of a card in the spaced repetition system
type CardState string

const (
	// CardStateNew represents a new card that hasn't been studied yet
	CardStateNew CardState = "new"
	// CardStateLearn represents a card in the learning phase
	CardStateLearn CardState = "learn"
	// CardStateReview represents a card in the review phase
	CardStateReview CardState = "review"
	// CardStateRelearn represents a card that was forgotten and is being relearned
	CardStateRelearn CardState = "relearn"
)

// IsValid checks if the card state is valid
func (s CardState) IsValid() bool {
	return s == CardStateNew || s == CardStateLearn || s == CardStateReview || s == CardStateRelearn
}

// String returns the string representation of the card state
func (s CardState) String() string {
	return string(s)
}

// CanBeStudied checks if a card in this state can be studied
// New cards can always be studied, learning/relearning cards can be studied if due,
// and review cards can be studied if due
func (s CardState) CanBeStudied() bool {
	return s.IsValid()
}

// IsNew checks if the state is new
func (s CardState) IsNew() bool {
	return s == CardStateNew
}

// IsLearning checks if the state is learning
func (s CardState) IsLearning() bool {
	return s == CardStateLearn
}

// IsReview checks if the state is review
func (s CardState) IsReview() bool {
	return s == CardStateReview
}

// IsRelearning checks if the state is relearning
func (s CardState) IsRelearning() bool {
	return s == CardStateRelearn
}

