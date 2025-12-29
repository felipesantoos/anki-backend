package deckoptionspreset

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrNameRequired   = errors.New("name is required")
)

type DeckOptionsPresetBuilder struct {
	preset *DeckOptionsPreset
	errs   []error
}

func NewBuilder() *DeckOptionsPresetBuilder {
	return &DeckOptionsPresetBuilder{
		preset: &DeckOptionsPreset{},
		errs:   make([]error, 0),
	}
}

func (b *DeckOptionsPresetBuilder) WithID(id int64) *DeckOptionsPresetBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.preset.id = id
	return b
}

func (b *DeckOptionsPresetBuilder) WithUserID(userID int64) *DeckOptionsPresetBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.preset.userID = userID
	return b
}

func (b *DeckOptionsPresetBuilder) WithName(name string) *DeckOptionsPresetBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.preset.name = name
	return b
}

func (b *DeckOptionsPresetBuilder) WithOptionsJSON(optionsJSON string) *DeckOptionsPresetBuilder {
	b.preset.optionsJSON = optionsJSON
	return b
}

func (b *DeckOptionsPresetBuilder) WithCreatedAt(createdAt time.Time) *DeckOptionsPresetBuilder {
	b.preset.createdAt = createdAt
	return b
}

func (b *DeckOptionsPresetBuilder) WithUpdatedAt(updatedAt time.Time) *DeckOptionsPresetBuilder {
	b.preset.updatedAt = updatedAt
	return b
}

func (b *DeckOptionsPresetBuilder) WithDeletedAt(deletedAt *time.Time) *DeckOptionsPresetBuilder {
	b.preset.deletedAt = deletedAt
	return b
}

func (b *DeckOptionsPresetBuilder) Build() (*DeckOptionsPreset, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.preset, nil
}

func (b *DeckOptionsPresetBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *DeckOptionsPresetBuilder) Errors() []error {
	return b.errs
}

