package deletionlog

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired  = errors.New("userID is required")
	ErrObjectTypeRequired = errors.New("objectType is required")
	ErrInvalidObjectType = errors.New("invalid objectType")
)

type DeletionLogBuilder struct {
	deletionLog *DeletionLog
	errs        []error
}

func NewBuilder() *DeletionLogBuilder {
	return &DeletionLogBuilder{
		deletionLog: &DeletionLog{},
		errs:        make([]error, 0),
	}
}

func (b *DeletionLogBuilder) WithID(id int64) *DeletionLogBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.deletionLog.id = id
	return b
}

func (b *DeletionLogBuilder) WithUserID(userID int64) *DeletionLogBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.deletionLog.userID = userID
	return b
}

func (b *DeletionLogBuilder) WithObjectType(objectType string) *DeletionLogBuilder {
	if objectType == "" {
		b.errs = append(b.errs, ErrObjectTypeRequired)
		return b
	}
	validTypes := map[string]bool{
		ObjectTypeNote:     true,
		ObjectTypeCard:     true,
		ObjectTypeDeck:     true,
		ObjectTypeNoteType: true,
	}
	if !validTypes[objectType] {
		b.errs = append(b.errs, ErrInvalidObjectType)
		return b
	}
	b.deletionLog.objectType = objectType
	return b
}

func (b *DeletionLogBuilder) WithObjectID(objectID int64) *DeletionLogBuilder {
	b.deletionLog.objectID = objectID
	return b
}

func (b *DeletionLogBuilder) WithObjectData(objectData string) *DeletionLogBuilder {
	b.deletionLog.objectData = objectData
	return b
}

func (b *DeletionLogBuilder) WithDeletedAt(deletedAt time.Time) *DeletionLogBuilder {
	b.deletionLog.deletedAt = deletedAt
	return b
}

func (b *DeletionLogBuilder) Build() (*DeletionLog, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.deletionLog, nil
}

func (b *DeletionLogBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *DeletionLogBuilder) Errors() []error {
	return b.errs
}

