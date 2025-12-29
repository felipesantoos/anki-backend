package media

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired  = errors.New("userID is required")
	ErrFilenameRequired = errors.New("filename is required")
)

type MediaBuilder struct {
	media *Media
	errs  []error // Lista de erros acumulados
}

func NewBuilder() *MediaBuilder {
	return &MediaBuilder{
		media: &Media{},
		errs:  make([]error, 0),
	}
}

func (b *MediaBuilder) WithID(id int64) *MediaBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.media.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *MediaBuilder) WithUserID(userID int64) *MediaBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.media.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) WithFilename(filename string) *MediaBuilder {
	if filename == "" {
		b.errs = append(b.errs, ErrFilenameRequired)
		return b
	}
	b.media.filename = filename // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) WithHash(hash string) *MediaBuilder {
	b.media.hash = hash // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) WithSize(size int64) *MediaBuilder {
	b.media.size = size // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) WithMimeType(mimeType string) *MediaBuilder {
	b.media.mimeType = mimeType // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) WithStoragePath(storagePath string) *MediaBuilder {
	b.media.storagePath = storagePath // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) WithCreatedAt(createdAt time.Time) *MediaBuilder {
	b.media.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) WithDeletedAt(deletedAt *time.Time) *MediaBuilder {
	b.media.deletedAt = deletedAt // Acesso direto ao campo privado
	return b
}

func (b *MediaBuilder) Build() (*Media, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.media, nil
}

// HasErrors retorna true se há erros acumulados
func (b *MediaBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *MediaBuilder) Errors() []error {
	return b.errs
}

