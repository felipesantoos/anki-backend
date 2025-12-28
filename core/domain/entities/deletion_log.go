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
	id         int64
	userID     int64
	objectType string // note, card, deck, note_type
	objectID   int64
	objectData string // JSONB in database
	deletedAt  time.Time
}

// Getters
func (dl *DeletionLog) GetID() int64 {
	return dl.id
}

func (dl *DeletionLog) GetUserID() int64 {
	return dl.userID
}

func (dl *DeletionLog) GetObjectID() int64 {
	return dl.objectID
}

func (dl *DeletionLog) GetObjectData() string {
	return dl.objectData
}

func (dl *DeletionLog) GetDeletedAt() time.Time {
	return dl.deletedAt
}

// Setters
func (dl *DeletionLog) SetID(id int64) {
	dl.id = id
}

func (dl *DeletionLog) SetUserID(userID int64) {
	dl.userID = userID
}

func (dl *DeletionLog) SetObjectType(objectType string) {
	dl.objectType = objectType
}

func (dl *DeletionLog) SetObjectID(objectID int64) {
	dl.objectID = objectID
}

func (dl *DeletionLog) SetObjectData(objectData string) {
	dl.objectData = objectData
}

func (dl *DeletionLog) SetDeletedAt(deletedAt time.Time) {
	dl.deletedAt = deletedAt
}

// GetObjectType returns the object type
func (dl *DeletionLog) GetObjectType() string {
	return dl.objectType
}

// CanRecover checks if the object can be recovered
// This is a domain method - actual recovery logic should be in service layer
func (dl *DeletionLog) CanRecover() bool {
	return dl.objectData != ""
}

