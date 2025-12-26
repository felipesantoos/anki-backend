package events

import (
	"time"
)

// DomainEvent represents a domain event that occurred in the system
// All domain events must implement this interface
type DomainEvent interface {
	// EventType returns the type/name of the event (e.g., "card.reviewed", "note.created")
	EventType() string

	// AggregateID returns the ID of the aggregate root that generated the event
	AggregateID() string

	// OccurredAt returns when the event occurred
	OccurredAt() time.Time

	// Metadata returns additional metadata about the event (can be nil or empty)
	// Common metadata includes: user_id, request_id, correlation_id, etc.
	Metadata() map[string]interface{}
}
