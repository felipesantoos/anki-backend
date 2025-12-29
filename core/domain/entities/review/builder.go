package review

import (
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrCardIDRequired   = errors.New("cardID is required")
	ErrInvalidRating    = errors.New("rating must be between 1 and 4")
	ErrInvalidReviewType = errors.New("invalid review type")
)

type ReviewBuilder struct {
	review *Review
	errs   []error // Lista de erros acumulados
}

func NewBuilder() *ReviewBuilder {
	return &ReviewBuilder{
		review: &Review{},
		errs:   make([]error, 0),
	}
}

func (b *ReviewBuilder) WithID(id int64) *ReviewBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.review.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *ReviewBuilder) WithCardID(cardID int64) *ReviewBuilder {
	if cardID <= 0 {
		b.errs = append(b.errs, ErrCardIDRequired)
		return b
	}
	b.review.cardID = cardID // Acesso direto ao campo privado
	return b
}

func (b *ReviewBuilder) WithRating(rating int) *ReviewBuilder {
	if rating < 1 || rating > 4 {
		b.errs = append(b.errs, ErrInvalidRating)
		return b
	}
	b.review.rating = rating // Acesso direto ao campo privado
	return b
}

func (b *ReviewBuilder) WithInterval(interval int) *ReviewBuilder {
	b.review.interval = interval // Acesso direto ao campo privado
	return b
}

func (b *ReviewBuilder) WithEase(ease int) *ReviewBuilder {
	b.review.ease = ease // Acesso direto ao campo privado
	return b
}

func (b *ReviewBuilder) WithTimeMs(timeMs int) *ReviewBuilder {
	b.review.timeMs = timeMs // Acesso direto ao campo privado
	return b
}

func (b *ReviewBuilder) WithType(reviewType valueobjects.ReviewType) *ReviewBuilder {
	if !reviewType.IsValid() {
		b.errs = append(b.errs, ErrInvalidReviewType)
		return b
	}
	b.review.reviewType = reviewType // Acesso direto ao campo privado
	return b
}

func (b *ReviewBuilder) WithCreatedAt(createdAt time.Time) *ReviewBuilder {
	b.review.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *ReviewBuilder) Build() (*Review, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.review, nil
}

// HasErrors retorna true se há erros acumulados
func (b *ReviewBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *ReviewBuilder) Errors() []error {
	return b.errs
}

