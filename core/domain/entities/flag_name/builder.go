package flagname

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired    = errors.New("userID is required")
	ErrInvalidFlagNumber = errors.New("flagNumber must be between 1 and 7")
	ErrNameRequired      = errors.New("name is required")
)

type FlagNameBuilder struct {
	flagName *FlagName
	errs     []error
}

func NewBuilder() *FlagNameBuilder {
	return &FlagNameBuilder{
		flagName: &FlagName{},
		errs:     make([]error, 0),
	}
}

func (b *FlagNameBuilder) WithID(id int64) *FlagNameBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.flagName.id = id
	return b
}

func (b *FlagNameBuilder) WithUserID(userID int64) *FlagNameBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.flagName.userID = userID
	return b
}

func (b *FlagNameBuilder) WithFlagNumber(flagNumber int) *FlagNameBuilder {
	if flagNumber < 1 || flagNumber > 7 {
		b.errs = append(b.errs, ErrInvalidFlagNumber)
		return b
	}
	b.flagName.flagNumber = flagNumber
	return b
}

func (b *FlagNameBuilder) WithName(name string) *FlagNameBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.flagName.name = name
	return b
}

func (b *FlagNameBuilder) WithCreatedAt(createdAt time.Time) *FlagNameBuilder {
	b.flagName.createdAt = createdAt
	return b
}

func (b *FlagNameBuilder) WithUpdatedAt(updatedAt time.Time) *FlagNameBuilder {
	b.flagName.updatedAt = updatedAt
	return b
}

func (b *FlagNameBuilder) Build() (*FlagName, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.flagName, nil
}

func (b *FlagNameBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *FlagNameBuilder) Errors() []error {
	return b.errs
}

