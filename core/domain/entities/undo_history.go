package entities

import (
	"time"
)

// OperationType represents the type of operation
const (
	OperationTypeEditNote    = "edit_note"
	OperationTypeDeleteNote  = "delete_note"
	OperationTypeMoveCard    = "move_card"
	OperationTypeChangeFlag  = "change_flag"
	OperationTypeAddTag      = "add_tag"
	OperationTypeRemoveTag   = "remove_tag"
	OperationTypeChangeDeck  = "change_deck"
)

// UndoHistory represents an undo history entry entity in the domain
// It stores operation history for undo/redo functionality
type UndoHistory struct {
	id            int64
	userID        int64
	operationType string // edit_note, delete_note, move_card, etc.
	operationData string // JSONB in database
	createdAt     time.Time
}

// Getters
func (uh *UndoHistory) GetID() int64 {
	return uh.id
}

func (uh *UndoHistory) GetUserID() int64 {
	return uh.userID
}

func (uh *UndoHistory) GetOperationData() string {
	return uh.operationData
}

func (uh *UndoHistory) GetCreatedAt() time.Time {
	return uh.createdAt
}

// Setters
func (uh *UndoHistory) SetID(id int64) {
	uh.id = id
}

func (uh *UndoHistory) SetUserID(userID int64) {
	uh.userID = userID
}

func (uh *UndoHistory) SetOperationType(operationType string) {
	uh.operationType = operationType
}

func (uh *UndoHistory) SetOperationData(operationData string) {
	uh.operationData = operationData
}

func (uh *UndoHistory) SetCreatedAt(createdAt time.Time) {
	uh.createdAt = createdAt
}

// GetOperationType returns the operation type
func (uh *UndoHistory) GetOperationType() string {
	return uh.operationType
}

// CanUndo checks if the operation can be undone
// This is a domain method - actual undo logic should be in service layer
func (uh *UndoHistory) CanUndo() bool {
	return uh.operationData != ""
}

