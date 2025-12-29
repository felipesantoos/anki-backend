package shareddeck

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrAuthorIDRequired = errors.New("authorID is required")
	ErrNameRequired     = errors.New("name is required")
)

type SharedDeckBuilder struct {
	sharedDeck *SharedDeck
	errs       []error
}

func NewBuilder() *SharedDeckBuilder {
	return &SharedDeckBuilder{
		sharedDeck: &SharedDeck{},
		errs:       make([]error, 0),
	}
}

func (b *SharedDeckBuilder) WithID(id int64) *SharedDeckBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.sharedDeck.id = id
	return b
}

func (b *SharedDeckBuilder) WithAuthorID(authorID int64) *SharedDeckBuilder {
	if authorID <= 0 {
		b.errs = append(b.errs, ErrAuthorIDRequired)
		return b
	}
	b.sharedDeck.authorID = authorID
	return b
}

func (b *SharedDeckBuilder) WithName(name string) *SharedDeckBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.sharedDeck.name = name
	return b
}

func (b *SharedDeckBuilder) WithDescription(description *string) *SharedDeckBuilder {
	b.sharedDeck.description = description
	return b
}

func (b *SharedDeckBuilder) WithCategory(category *string) *SharedDeckBuilder {
	b.sharedDeck.category = category
	return b
}

func (b *SharedDeckBuilder) WithPackagePath(packagePath string) *SharedDeckBuilder {
	b.sharedDeck.packagePath = packagePath
	return b
}

func (b *SharedDeckBuilder) WithPackageSize(packageSize int64) *SharedDeckBuilder {
	b.sharedDeck.packageSize = packageSize
	return b
}

func (b *SharedDeckBuilder) WithDownloadCount(downloadCount int) *SharedDeckBuilder {
	b.sharedDeck.downloadCount = downloadCount
	return b
}

func (b *SharedDeckBuilder) WithRatingAverage(ratingAverage float64) *SharedDeckBuilder {
	b.sharedDeck.ratingAverage = ratingAverage
	return b
}

func (b *SharedDeckBuilder) WithRatingCount(ratingCount int) *SharedDeckBuilder {
	b.sharedDeck.ratingCount = ratingCount
	return b
}

func (b *SharedDeckBuilder) WithTags(tags []string) *SharedDeckBuilder {
	b.sharedDeck.tags = tags
	return b
}

func (b *SharedDeckBuilder) WithIsFeatured(isFeatured bool) *SharedDeckBuilder {
	b.sharedDeck.isFeatured = isFeatured
	return b
}

func (b *SharedDeckBuilder) WithIsPublic(isPublic bool) *SharedDeckBuilder {
	b.sharedDeck.isPublic = isPublic
	return b
}

func (b *SharedDeckBuilder) WithCreatedAt(createdAt time.Time) *SharedDeckBuilder {
	b.sharedDeck.createdAt = createdAt
	return b
}

func (b *SharedDeckBuilder) WithUpdatedAt(updatedAt time.Time) *SharedDeckBuilder {
	b.sharedDeck.updatedAt = updatedAt
	return b
}

func (b *SharedDeckBuilder) WithDeletedAt(deletedAt *time.Time) *SharedDeckBuilder {
	b.sharedDeck.deletedAt = deletedAt
	return b
}

func (b *SharedDeckBuilder) Build() (*SharedDeck, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.sharedDeck, nil
}

func (b *SharedDeckBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *SharedDeckBuilder) Errors() []error {
	return b.errs
}

