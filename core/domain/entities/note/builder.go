package note

import (
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrUserIDRequired    = errors.New("userID is required")
	ErrNoteTypeIDRequired = errors.New("noteTypeID is required")
	ErrGUIDRequired       = errors.New("guid is required")
)

type NoteBuilder struct {
	note *Note
	errs []error // Lista de erros acumulados
}

func NewBuilder() *NoteBuilder {
	return &NoteBuilder{
		note: &Note{},
		errs: make([]error, 0),
	}
}

func (b *NoteBuilder) WithID(id int64) *NoteBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.note.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *NoteBuilder) WithUserID(userID int64) *NoteBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.note.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithGUID(guid valueobjects.GUID) *NoteBuilder {
	if guid.Value() == "" || guid.IsEmpty() {
		b.errs = append(b.errs, ErrGUIDRequired)
		return b
	}
	b.note.guid = guid // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithNoteTypeID(noteTypeID int64) *NoteBuilder {
	if noteTypeID <= 0 {
		b.errs = append(b.errs, ErrNoteTypeIDRequired)
		return b
	}
	b.note.noteTypeID = noteTypeID // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithFieldsJSON(fieldsJSON string) *NoteBuilder {
	b.note.fieldsJSON = fieldsJSON // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithTags(tags []string) *NoteBuilder {
	b.note.tags = tags // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithMarked(marked bool) *NoteBuilder {
	b.note.marked = marked // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithCreatedAt(createdAt time.Time) *NoteBuilder {
	b.note.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithUpdatedAt(updatedAt time.Time) *NoteBuilder {
	b.note.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) WithDeletedAt(deletedAt *time.Time) *NoteBuilder {
	b.note.deletedAt = deletedAt // Acesso direto ao campo privado
	return b
}

func (b *NoteBuilder) Build() (*Note, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.note, nil
}

// HasErrors retorna true se há erros acumulados
func (b *NoteBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *NoteBuilder) Errors() []error {
	return b.errs
}

