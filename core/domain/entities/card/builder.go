package card

import (
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrNoteIDRequired = errors.New("noteID is required")
	ErrDeckIDRequired = errors.New("deckID is required")
	ErrInvalidState   = errors.New("invalid card state")
)

type CardBuilder struct {
	card *Card
	errs []error // Lista de erros acumulados
}

func NewBuilder() *CardBuilder {
	return &CardBuilder{
		card: &Card{},
		errs: make([]error, 0),
	}
}

func (b *CardBuilder) WithID(id int64) *CardBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.card.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *CardBuilder) WithNoteID(noteID int64) *CardBuilder {
	if noteID <= 0 {
		b.errs = append(b.errs, ErrNoteIDRequired)
		return b
	}
	b.card.noteID = noteID // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithCardTypeID(cardTypeID int) *CardBuilder {
	b.card.cardTypeID = cardTypeID // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithDeckID(deckID int64) *CardBuilder {
	if deckID <= 0 {
		b.errs = append(b.errs, ErrDeckIDRequired)
		return b
	}
	b.card.deckID = deckID // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithHomeDeckID(homeDeckID *int64) *CardBuilder {
	b.card.homeDeckID = homeDeckID // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithDue(due int64) *CardBuilder {
	b.card.due = due // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithInterval(interval int) *CardBuilder {
	b.card.interval = interval // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithEase(ease int) *CardBuilder {
	b.card.ease = ease // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithLapses(lapses int) *CardBuilder {
	b.card.lapses = lapses // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithReps(reps int) *CardBuilder {
	b.card.reps = reps // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithState(state valueobjects.CardState) *CardBuilder {
	if !state.IsValid() {
		b.errs = append(b.errs, ErrInvalidState)
		return b
	}
	b.card.state = state // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithPosition(position int) *CardBuilder {
	b.card.position = position // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithFlag(flag int) *CardBuilder {
	if flag < 0 || flag > 7 {
		b.errs = append(b.errs, ErrInvalidFlag)
		return b
	}
	b.card.flag = flag // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithSuspended(suspended bool) *CardBuilder {
	b.card.suspended = suspended // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithBuried(buried bool) *CardBuilder {
	b.card.buried = buried // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithStability(stability *float64) *CardBuilder {
	b.card.stability = stability // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithDifficulty(difficulty *float64) *CardBuilder {
	b.card.difficulty = difficulty // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithLastReviewAt(lastReviewAt *time.Time) *CardBuilder {
	b.card.lastReviewAt = lastReviewAt // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithCreatedAt(createdAt time.Time) *CardBuilder {
	b.card.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) WithUpdatedAt(updatedAt time.Time) *CardBuilder {
	b.card.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *CardBuilder) Build() (*Card, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.card, nil
}

// HasErrors retorna true se há erros acumulados
func (b *CardBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *CardBuilder) Errors() []error {
	return b.errs
}

