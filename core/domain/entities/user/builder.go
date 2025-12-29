package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrInvalidEmail     = errors.New("invalid email format")
)

type UserBuilder struct {
	user *User
	errs []error // Lista de erros acumulados
}

func NewBuilder() *UserBuilder {
	return &UserBuilder{
		user: &User{},
		errs: make([]error, 0),
	}
}

func (b *UserBuilder) WithID(id int64) *UserBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.user.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *UserBuilder) WithEmail(email valueobjects.Email) *UserBuilder {
	// Validação no método With...
	if email.Value() == "" {
		b.errs = append(b.errs, ErrEmailRequired)
		return b
	}
	b.user.email = email // Acesso direto ao campo privado
	return b
}

func (b *UserBuilder) WithPasswordHash(passwordHash valueobjects.Password) *UserBuilder {
	// Validação no método With...
	if passwordHash.Hash() == "" {
		b.errs = append(b.errs, ErrPasswordRequired)
		return b
	}
	b.user.passwordHash = passwordHash // Acesso direto ao campo privado
	return b
}

func (b *UserBuilder) WithEmailVerified(verified bool) *UserBuilder {
	b.user.emailVerified = verified // Acesso direto ao campo privado
	return b
}

func (b *UserBuilder) WithCreatedAt(createdAt time.Time) *UserBuilder {
	b.user.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *UserBuilder) WithUpdatedAt(updatedAt time.Time) *UserBuilder {
	b.user.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *UserBuilder) WithLastLoginAt(lastLoginAt *time.Time) *UserBuilder {
	b.user.lastLoginAt = lastLoginAt // Acesso direto ao campo privado
	return b
}

func (b *UserBuilder) WithDeletedAt(deletedAt *time.Time) *UserBuilder {
	b.user.deletedAt = deletedAt // Acesso direto ao campo privado
	return b
}

func (b *UserBuilder) Build() (*User, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.user, nil
}

// HasErrors retorna true se há erros acumulados
func (b *UserBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *UserBuilder) Errors() []error {
	return b.errs
}

