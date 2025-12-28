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
	ID            int64
	UserID        int64
	OperationType string // edit_note, delete_note, move_card, etc.
	OperationData string // JSONB in database
	CreatedAt     time.Time
}

// GetOperationType returns the operation type
func (uh *UndoHistory) GetOperationType() string {
	return uh.OperationType
}

// CanUndo checks if the operation can be undone
// This is a domain method - actual undo logic should be in service layer
func (uh *UndoHistory) CanUndo() bool {
	return uh.OperationData != ""
}

