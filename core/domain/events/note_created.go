package events

import (
	"strconv"
	"time"
)

// NoteCreatedEventType is the event type constant for NoteCreated events
const NoteCreatedEventType = "note.created"

// NoteCreated is published when a note is created
type NoteCreated struct {
	NoteID    int64
	UserID    int64
	NoteTypeID int64
	Timestamp time.Time
}

// EventType returns the type of the event
func (e *NoteCreated) EventType() string {
	return NoteCreatedEventType
}

// AggregateID returns the note ID as the aggregate root ID
func (e *NoteCreated) AggregateID() string {
	return strconv.FormatInt(e.NoteID, 10)
}

// OccurredAt returns when the event occurred
func (e *NoteCreated) OccurredAt() time.Time {
	return e.Timestamp
}

// Metadata returns additional metadata about the event
func (e *NoteCreated) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"user_id":    e.UserID,
		"note_type_id": e.NoteTypeID,
	}
}
