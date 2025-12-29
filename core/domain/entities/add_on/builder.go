package addon

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrCodeRequired   = errors.New("code is required")
	ErrNameRequired   = errors.New("name is required")
)

type AddOnBuilder struct {
	addOn *AddOn
	errs  []error // Lista de erros acumulados
}

func NewBuilder() *AddOnBuilder {
	return &AddOnBuilder{
		addOn: &AddOn{},
		errs:  make([]error, 0),
	}
}

func (b *AddOnBuilder) WithID(id int64) *AddOnBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.addOn.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *AddOnBuilder) WithUserID(userID int64) *AddOnBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.addOn.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) WithCode(code string) *AddOnBuilder {
	if code == "" {
		b.errs = append(b.errs, ErrCodeRequired)
		return b
	}
	b.addOn.code = code // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) WithName(name string) *AddOnBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.addOn.name = name // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) WithVersion(version string) *AddOnBuilder {
	b.addOn.version = version // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) WithEnabled(enabled bool) *AddOnBuilder {
	b.addOn.enabled = enabled // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) WithConfigJSON(configJSON string) *AddOnBuilder {
	b.addOn.configJSON = configJSON // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) WithInstalledAt(installedAt time.Time) *AddOnBuilder {
	b.addOn.installedAt = installedAt // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) WithUpdatedAt(updatedAt time.Time) *AddOnBuilder {
	b.addOn.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *AddOnBuilder) Build() (*AddOn, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.addOn, nil
}

// HasErrors retorna true se há erros acumulados
func (b *AddOnBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *AddOnBuilder) Errors() []error {
	return b.errs
}

