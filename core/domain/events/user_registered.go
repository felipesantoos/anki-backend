package events

import (
	"strconv"
	"time"
)

// UserRegisteredEventType is the event type constant for UserRegistered events
const UserRegisteredEventType = "user.registered"

// UserRegistered is published when a new user registers in the system
type UserRegistered struct {
	UserID    int64
	Email     string
	Timestamp time.Time
}

// EventType returns the type of the event
func (e *UserRegistered) EventType() string {
	return UserRegisteredEventType
}

// AggregateID returns the user ID as the aggregate root ID
func (e *UserRegistered) AggregateID() string {
	return strconv.FormatInt(e.UserID, 10)
}

// OccurredAt returns when the event occurred
func (e *UserRegistered) OccurredAt() time.Time {
	return e.Timestamp
}

// Metadata returns additional metadata about the event
func (e *UserRegistered) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"email": e.Email,
	}
}
