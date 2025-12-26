package events

import (
	"strconv"
	"time"
)

// CardReviewedEventType is the event type constant for CardReviewed events
const CardReviewedEventType = "card.reviewed"

// CardReviewed is published when a card is reviewed
type CardReviewed struct {
	CardID    int64
	UserID    int64
	Rating    int
	NewState  string // Using string for now, can be changed to valueobjects.CardState when available
	Timestamp time.Time
}

// EventType returns the type of the event
func (e *CardReviewed) EventType() string {
	return CardReviewedEventType
}

// AggregateID returns the card ID as the aggregate root ID
func (e *CardReviewed) AggregateID() string {
	return strconv.FormatInt(e.CardID, 10)
}

// OccurredAt returns when the event occurred
func (e *CardReviewed) OccurredAt() time.Time {
	return e.Timestamp
}

// Metadata returns additional metadata about the event
func (e *CardReviewed) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"user_id":  e.UserID,
		"rating":   e.Rating,
		"new_state": e.NewState,
	}
}
