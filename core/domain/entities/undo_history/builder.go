package undohistory

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired      = errors.New("userID is required")
	ErrOperationTypeRequired = errors.New("operationType is required")
	ErrInvalidOperationType = errors.New("invalid operationType")
)

type UndoHistoryBuilder struct {
	undoHistory *UndoHistory
	errs        []error
}

func NewBuilder() *UndoHistoryBuilder {
	return &UndoHistoryBuilder{
		undoHistory: &UndoHistory{},
		errs:        make([]error, 0),
	}
}

func (b *UndoHistoryBuilder) WithID(id int64) *UndoHistoryBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.undoHistory.id = id
	return b
}

func (b *UndoHistoryBuilder) WithUserID(userID int64) *UndoHistoryBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.undoHistory.userID = userID
	return b
}

func (b *UndoHistoryBuilder) WithOperationType(operationType string) *UndoHistoryBuilder {
	if operationType == "" {
		b.errs = append(b.errs, ErrOperationTypeRequired)
		return b
	}
	validTypes := map[string]bool{
		OperationTypeEditNote:   true,
		OperationTypeDeleteNote: true,
		OperationTypeMoveCard:   true,
		OperationTypeChangeFlag: true,
		OperationTypeAddTag:     true,
		OperationTypeRemoveTag:  true,
		OperationTypeChangeDeck: true,
	}
	if !validTypes[operationType] {
		b.errs = append(b.errs, ErrInvalidOperationType)
		return b
	}
	b.undoHistory.operationType = operationType
	return b
}

func (b *UndoHistoryBuilder) WithOperationData(operationData string) *UndoHistoryBuilder {
	b.undoHistory.operationData = operationData
	return b
}

func (b *UndoHistoryBuilder) WithCreatedAt(createdAt time.Time) *UndoHistoryBuilder {
	b.undoHistory.createdAt = createdAt
	return b
}

func (b *UndoHistoryBuilder) Build() (*UndoHistory, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.undoHistory, nil
}

func (b *UndoHistoryBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *UndoHistoryBuilder) Errors() []error {
	return b.errs
}

