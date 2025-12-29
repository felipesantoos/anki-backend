package browserconfig

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
)

type BrowserConfigBuilder struct {
	browserConfig *BrowserConfig
	errs          []error
}

func NewBuilder() *BrowserConfigBuilder {
	return &BrowserConfigBuilder{
		browserConfig: &BrowserConfig{},
		errs:          make([]error, 0),
	}
}

func (b *BrowserConfigBuilder) WithID(id int64) *BrowserConfigBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.browserConfig.id = id
	return b
}

func (b *BrowserConfigBuilder) WithUserID(userID int64) *BrowserConfigBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.browserConfig.userID = userID
	return b
}

func (b *BrowserConfigBuilder) WithVisibleColumns(visibleColumns []string) *BrowserConfigBuilder {
	b.browserConfig.visibleColumns = visibleColumns
	return b
}

func (b *BrowserConfigBuilder) WithColumnWidths(columnWidths string) *BrowserConfigBuilder {
	b.browserConfig.columnWidths = columnWidths
	return b
}

func (b *BrowserConfigBuilder) WithSortColumn(sortColumn *string) *BrowserConfigBuilder {
	b.browserConfig.sortColumn = sortColumn
	return b
}

func (b *BrowserConfigBuilder) WithSortDirection(sortDirection string) *BrowserConfigBuilder {
	b.browserConfig.sortDirection = sortDirection
	return b
}

func (b *BrowserConfigBuilder) WithCreatedAt(createdAt time.Time) *BrowserConfigBuilder {
	b.browserConfig.createdAt = createdAt
	return b
}

func (b *BrowserConfigBuilder) WithUpdatedAt(updatedAt time.Time) *BrowserConfigBuilder {
	b.browserConfig.updatedAt = updatedAt
	return b
}

func (b *BrowserConfigBuilder) Build() (*BrowserConfig, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.browserConfig, nil
}

func (b *BrowserConfigBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *BrowserConfigBuilder) Errors() []error {
	return b.errs
}

