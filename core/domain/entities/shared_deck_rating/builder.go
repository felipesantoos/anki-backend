package shareddeckrating

import (
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrSharedDeckIDRequired = errors.New("sharedDeckID is required")
	ErrUserIDRequired       = errors.New("userID is required")
	ErrInvalidRating        = errors.New("invalid rating")
)

type SharedDeckRatingBuilder struct {
	sharedDeckRating *SharedDeckRating
	errs             []error
}

func NewBuilder() *SharedDeckRatingBuilder {
	return &SharedDeckRatingBuilder{
		sharedDeckRating: &SharedDeckRating{},
		errs:             make([]error, 0),
	}
}

func (b *SharedDeckRatingBuilder) WithID(id int64) *SharedDeckRatingBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.sharedDeckRating.id = id
	return b
}

func (b *SharedDeckRatingBuilder) WithSharedDeckID(sharedDeckID int64) *SharedDeckRatingBuilder {
	if sharedDeckID <= 0 {
		b.errs = append(b.errs, ErrSharedDeckIDRequired)
		return b
	}
	b.sharedDeckRating.sharedDeckID = sharedDeckID
	return b
}

func (b *SharedDeckRatingBuilder) WithUserID(userID int64) *SharedDeckRatingBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.sharedDeckRating.userID = userID
	return b
}

func (b *SharedDeckRatingBuilder) WithRating(rating valueobjects.Rating) *SharedDeckRatingBuilder {
	if !rating.IsValid() {
		b.errs = append(b.errs, ErrInvalidRating)
		return b
	}
	b.sharedDeckRating.rating = rating
	return b
}

func (b *SharedDeckRatingBuilder) WithComment(comment *string) *SharedDeckRatingBuilder {
	b.sharedDeckRating.comment = comment
	return b
}

func (b *SharedDeckRatingBuilder) WithCreatedAt(createdAt time.Time) *SharedDeckRatingBuilder {
	b.sharedDeckRating.createdAt = createdAt
	return b
}

func (b *SharedDeckRatingBuilder) WithUpdatedAt(updatedAt time.Time) *SharedDeckRatingBuilder {
	b.sharedDeckRating.updatedAt = updatedAt
	return b
}

func (b *SharedDeckRatingBuilder) Build() (*SharedDeckRating, error) {
	if len(b.errs) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.sharedDeckRating, nil
}

func (b *SharedDeckRatingBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

func (b *SharedDeckRatingBuilder) Errors() []error {
	return b.errs
}

