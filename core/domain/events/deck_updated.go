package events

import (
	"strconv"
	"time"
)

// DeckUpdatedEventType is the event type constant for DeckUpdated events
const DeckUpdatedEventType = "deck.updated"

// DeckUpdated is published when a deck is updated
type DeckUpdated struct {
	DeckID    int64
	UserID    int64
	Timestamp time.Time
}

// EventType returns the type of the event
func (e *DeckUpdated) EventType() string {
	return DeckUpdatedEventType
}

// AggregateID returns the deck ID as the aggregate root ID
func (e *DeckUpdated) AggregateID() string {
	return strconv.FormatInt(e.DeckID, 10)
}

// OccurredAt returns when the event occurred
func (e *DeckUpdated) OccurredAt() time.Time {
	return e.Timestamp
}

// Metadata returns additional metadata about the event
func (e *DeckUpdated) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"user_id": e.UserID,
	}
}
