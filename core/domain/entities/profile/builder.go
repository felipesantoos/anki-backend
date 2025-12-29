package profile

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrNameRequired   = errors.New("name is required")
)

type ProfileBuilder struct {
	profile *Profile
	errs    []error // Lista de erros acumulados
}

func NewBuilder() *ProfileBuilder {
	return &ProfileBuilder{
		profile: &Profile{},
		errs:    make([]error, 0),
	}
}

func (b *ProfileBuilder) WithID(id int64) *ProfileBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.profile.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *ProfileBuilder) WithUserID(userID int64) *ProfileBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.profile.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *ProfileBuilder) WithName(name string) *ProfileBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.profile.name = name // Acesso direto ao campo privado
	return b
}

func (b *ProfileBuilder) WithAnkiWebSyncEnabled(ankiWebSyncEnabled bool) *ProfileBuilder {
	b.profile.ankiWebSyncEnabled = ankiWebSyncEnabled // Acesso direto ao campo privado
	return b
}

func (b *ProfileBuilder) WithAnkiWebUsername(ankiWebUsername *string) *ProfileBuilder {
	b.profile.ankiWebUsername = ankiWebUsername // Acesso direto ao campo privado
	return b
}

func (b *ProfileBuilder) WithCreatedAt(createdAt time.Time) *ProfileBuilder {
	b.profile.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *ProfileBuilder) WithUpdatedAt(updatedAt time.Time) *ProfileBuilder {
	b.profile.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *ProfileBuilder) WithDeletedAt(deletedAt *time.Time) *ProfileBuilder {
	b.profile.deletedAt = deletedAt // Acesso direto ao campo privado
	return b
}

func (b *ProfileBuilder) Build() (*Profile, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.profile, nil
}

// HasErrors retorna true se há erros acumulados
func (b *ProfileBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *ProfileBuilder) Errors() []error {
	return b.errs
}

