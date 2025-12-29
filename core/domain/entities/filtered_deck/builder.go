package filtereddeck

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrNameRequired   = errors.New("name is required")
)

type FilteredDeckBuilder struct {
	filteredDeck *FilteredDeck
	errs         []error
}

func NewBuilder() *FilteredDeckBuilder {
	return &FilteredDeckBuilder{
		filteredDeck: &FilteredDeck{},
		errs:         make([]error, 0),
	}
}

func (b *FilteredDeckBuilder) WithID(id int64) *FilteredDeckBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.filteredDeck.id = id
	return b
}

func (b *FilteredDeckBuilder) WithUserID(userID int64) *FilteredDeckBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.filteredDeck.userID = userID
	return b
}

func (b *FilteredDeckBuilder) WithName(name string) *FilteredDeckBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.filteredDeck.name = name
	return b
}

func (b *FilteredDeckBuilder) WithSearchFilter(searchFilter string) *FilteredDeckBuilder {
	b.filteredDeck.searchFilter = searchFilter
	return b
}

func (b *FilteredDeckBuilder) WithSecondFilter(secondFilter *string) *FilteredDeckBuilder {
	b.filteredDeck.secondFilter = secondFilter
	return b
}

func (b *FilteredDeckBuilder) WithLimitCards(limitCards int) *FilteredDeckBuilder {
	b.filteredDeck.limitCards = limitCards
	return b
}

func (b *FilteredDeckBuilder) WithOrderBy(orderBy string) *FilteredDeckBuilder {
	b.filteredDeck.orderBy = orderBy
	return b
}

func (b *FilteredDeckBuilder) WithReschedule(reschedule bool) *FilteredDeckBuilder {
	b.filteredDeck.reschedule = reschedule
	return b
}

func (b *FilteredDeckBuilder) WithCreatedAt(createdAt time.Time) *FilteredDeckBuilder {
	b.filteredDeck.createdAt = createdAt
	return b
}

func (b *FilteredDeckBuilder) WithUpdatedAt(updatedAt time.Time) *FilteredDeckBuilder {
	b.filteredDeck.updatedAt = updatedAt
	return b
}

func (b *FilteredDeckBuilder) WithLastRebuildAt(lastRebuildAt *time.Time) *FilteredDeckBuilder {
	b.filteredDeck.lastRebuildAt = lastRebuildAt
	return b
}

func (b *FilteredDeckBuilder) WithDeletedAt(deletedAt *time.Time) *FilteredDeckBuilder {
	b.filteredDeck.deletedAt = deletedAt
	return b
}

func (b *FilteredDeckBuilder) Build() (*FilteredDeck, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.filteredDeck, nil
}

func (b *FilteredDeckBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *FilteredDeckBuilder) Errors() []error {
	return b.errs
}

