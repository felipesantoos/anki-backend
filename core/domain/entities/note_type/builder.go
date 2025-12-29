package notetype

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrNameRequired   = errors.New("name is required")
)

type NoteTypeBuilder struct {
	noteType *NoteType
	errs     []error // Lista de erros acumulados
}

func NewBuilder() *NoteTypeBuilder {
	return &NoteTypeBuilder{
		noteType: &NoteType{},
		errs:     make([]error, 0),
	}
}

func (b *NoteTypeBuilder) WithID(id int64) *NoteTypeBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.noteType.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *NoteTypeBuilder) WithUserID(userID int64) *NoteTypeBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.noteType.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) WithName(name string) *NoteTypeBuilder {
	if name == "" {
		b.errs = append(b.errs, ErrNameRequired)
		return b
	}
	b.noteType.name = name // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) WithFieldsJSON(fieldsJSON string) *NoteTypeBuilder {
	b.noteType.fieldsJSON = fieldsJSON // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) WithCardTypesJSON(cardTypesJSON string) *NoteTypeBuilder {
	b.noteType.cardTypesJSON = cardTypesJSON // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) WithTemplatesJSON(templatesJSON string) *NoteTypeBuilder {
	b.noteType.templatesJSON = templatesJSON // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) WithCreatedAt(createdAt time.Time) *NoteTypeBuilder {
	b.noteType.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) WithUpdatedAt(updatedAt time.Time) *NoteTypeBuilder {
	b.noteType.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) WithDeletedAt(deletedAt *time.Time) *NoteTypeBuilder {
	b.noteType.deletedAt = deletedAt // Acesso direto ao campo privado
	return b
}

func (b *NoteTypeBuilder) Build() (*NoteType, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.noteType, nil
}

// HasErrors retorna true se há erros acumulados
func (b *NoteTypeBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *NoteTypeBuilder) Errors() []error {
	return b.errs
}

