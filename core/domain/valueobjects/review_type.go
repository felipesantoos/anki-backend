package valueobjects

// ReviewType represents the type of a review session
type ReviewType string

const (
	// ReviewTypeLearn represents a learning review
	ReviewTypeLearn ReviewType = "learn"
	// ReviewTypeReview represents a regular review
	ReviewTypeReview ReviewType = "review"
	// ReviewTypeRelearn represents a relearning review
	ReviewTypeRelearn ReviewType = "relearn"
	// ReviewTypeCram represents a cram session review
	ReviewTypeCram ReviewType = "cram"
)

// IsValid checks if the review type is valid
func (t ReviewType) IsValid() bool {
	return t == ReviewTypeLearn || t == ReviewTypeReview || t == ReviewTypeRelearn || t == ReviewTypeCram
}

// String returns the string representation of the review type
func (t ReviewType) String() string {
	return string(t)
}

