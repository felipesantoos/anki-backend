package deck

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrNameRequired   = errors.New("name is required")
)

type DeckBuilder struct {
	deck *Deck
	errs []error // Lista de erros acumulados
}

func NewBuilder() *DeckBuilder {
	return &DeckBuilder{
		deck: &Deck{},
		errs: make([]error, 0),
	}
}

func (b *DeckBuilder) WithID(id int64) *DeckBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.deck.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *DeckBuilder) WithUserID(userID int64) *DeckBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.deck.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *DeckBuilder) WithName(name string) *DeckBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.deck.name = name // Acesso direto ao campo privado
	return b
}

func (b *DeckBuilder) WithParentID(parentID *int64) *DeckBuilder {
	b.deck.parentID = parentID // Acesso direto ao campo privado
	return b
}

func (b *DeckBuilder) WithOptionsJSON(optionsJSON string) *DeckBuilder {
	b.deck.optionsJSON = optionsJSON // Acesso direto ao campo privado
	return b
}

func (b *DeckBuilder) WithCreatedAt(createdAt time.Time) *DeckBuilder {
	b.deck.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *DeckBuilder) WithUpdatedAt(updatedAt time.Time) *DeckBuilder {
	b.deck.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *DeckBuilder) WithDeletedAt(deletedAt *time.Time) *DeckBuilder {
	b.deck.deletedAt = deletedAt // Acesso direto ao campo privado
	return b
}

func (b *DeckBuilder) Build() (*Deck, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.deck, nil
}

// HasErrors retorna true se há erros acumulados
func (b *DeckBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *DeckBuilder) Errors() []error {
	return b.errs
}

