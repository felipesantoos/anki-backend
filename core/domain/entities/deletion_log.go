package entities

import (
	"time"
)

// ObjectType represents the type of deleted object
const (
	ObjectTypeNote     = "note"
	ObjectTypeCard     = "card"
	ObjectTypeDeck     = "deck"
	ObjectTypeNoteType = "note_type"
)

// DeletionLog represents a deletion log entry entity in the domain
// It stores information about deleted objects for potential recovery
type DeletionLog struct {
	ID         int64
	UserID     int64
	ObjectType string // note, card, deck, note_type
	ObjectID   int64
	ObjectData string // JSONB in database
	DeletedAt  time.Time
}

// GetObjectType returns the object type
func (dl *DeletionLog) GetObjectType() string {
	return dl.ObjectType
}

// CanRecover checks if the object can be recovered
// This is a domain method - actual recovery logic should be in service layer
func (dl *DeletionLog) CanRecover() bool {
	return dl.ObjectData != ""
}

