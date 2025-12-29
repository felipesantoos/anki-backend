package savedsearch

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrNameRequired   = errors.New("name is required")
)

type SavedSearchBuilder struct {
	savedSearch *SavedSearch
	errs        []error
}

func NewBuilder() *SavedSearchBuilder {
	return &SavedSearchBuilder{
		savedSearch: &SavedSearch{},
		errs:        make([]error, 0),
	}
}

func (b *SavedSearchBuilder) WithID(id int64) *SavedSearchBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.savedSearch.id = id
	return b
}

func (b *SavedSearchBuilder) WithUserID(userID int64) *SavedSearchBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.savedSearch.userID = userID
	return b
}

func (b *SavedSearchBuilder) WithName(name string) *SavedSearchBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.savedSearch.name = name
	return b
}

func (b *SavedSearchBuilder) WithSearchQuery(searchQuery string) *SavedSearchBuilder {
	b.savedSearch.searchQuery = searchQuery
	return b
}

func (b *SavedSearchBuilder) WithCreatedAt(createdAt time.Time) *SavedSearchBuilder {
	b.savedSearch.createdAt = createdAt
	return b
}

func (b *SavedSearchBuilder) WithUpdatedAt(updatedAt time.Time) *SavedSearchBuilder {
	b.savedSearch.updatedAt = updatedAt
	return b
}

func (b *SavedSearchBuilder) WithDeletedAt(deletedAt *time.Time) *SavedSearchBuilder {
	b.savedSearch.deletedAt = deletedAt
	return b
}

func (b *SavedSearchBuilder) Build() (*SavedSearch, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.savedSearch, nil
}

func (b *SavedSearchBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *SavedSearchBuilder) Errors() []error {
	return b.errs
}

