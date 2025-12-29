package syncmeta

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrClientIDRequired = errors.New("clientID is required")
)

type SyncMetaBuilder struct {
	syncMeta *SyncMeta
	errs     []error // Lista de erros acumulados
}

func NewBuilder() *SyncMetaBuilder {
	return &SyncMetaBuilder{
		syncMeta: &SyncMeta{},
		errs:     make([]error, 0),
	}
}

func (b *SyncMetaBuilder) WithID(id int64) *SyncMetaBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.syncMeta.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *SyncMetaBuilder) WithUserID(userID int64) *SyncMetaBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.syncMeta.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *SyncMetaBuilder) WithClientID(clientID string) *SyncMetaBuilder {
	if clientID == "" {
		b.errs = append(b.errs, ErrClientIDRequired)
		return b
	}
	b.syncMeta.clientID = clientID // Acesso direto ao campo privado
	return b
}

func (b *SyncMetaBuilder) WithLastSync(lastSync time.Time) *SyncMetaBuilder {
	b.syncMeta.lastSync = lastSync // Acesso direto ao campo privado
	return b
}

func (b *SyncMetaBuilder) WithLastSyncUSN(lastSyncUSN int64) *SyncMetaBuilder {
	b.syncMeta.lastSyncUSN = lastSyncUSN // Acesso direto ao campo privado
	return b
}

func (b *SyncMetaBuilder) WithCreatedAt(createdAt time.Time) *SyncMetaBuilder {
	b.syncMeta.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *SyncMetaBuilder) WithUpdatedAt(updatedAt time.Time) *SyncMetaBuilder {
	b.syncMeta.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *SyncMetaBuilder) Build() (*SyncMeta, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.syncMeta, nil
}

// HasErrors retorna true se há erros acumulados
func (b *SyncMetaBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *SyncMetaBuilder) Errors() []error {
	return b.errs
}

