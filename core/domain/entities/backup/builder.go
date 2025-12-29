package backup

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserIDRequired  = errors.New("userID is required")
	ErrFilenameRequired = errors.New("filename is required")
	ErrInvalidBackupType = errors.New("invalid backup type")
)

type BackupBuilder struct {
	backup *Backup
	errs   []error // Lista de erros acumulados
}

func NewBuilder() *BackupBuilder {
	return &BackupBuilder{
		backup: &Backup{},
		errs:   make([]error, 0),
	}
}

func (b *BackupBuilder) WithID(id int64) *BackupBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.backup.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *BackupBuilder) WithUserID(userID int64) *BackupBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.backup.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *BackupBuilder) WithFilename(filename string) *BackupBuilder {
	if filename == "" {
		b.errs = append(b.errs, ErrFilenameRequired)
		return b
	}
	b.backup.filename = filename // Acesso direto ao campo privado
	return b
}

func (b *BackupBuilder) WithSize(size int64) *BackupBuilder {
	b.backup.size = size // Acesso direto ao campo privado
	return b
}

func (b *BackupBuilder) WithStoragePath(storagePath string) *BackupBuilder {
	b.backup.storagePath = storagePath // Acesso direto ao campo privado
	return b
}

func (b *BackupBuilder) WithBackupType(backupType string) *BackupBuilder {
	if backupType != BackupTypeAutomatic && backupType != BackupTypeManual && backupType != BackupTypePreOperation {
		b.errs = append(b.errs, ErrInvalidBackupType)
		return b
	}
	b.backup.backupType = backupType // Acesso direto ao campo privado
	return b
}

func (b *BackupBuilder) WithCreatedAt(createdAt time.Time) *BackupBuilder {
	b.backup.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *BackupBuilder) Build() (*Backup, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.backup, nil
}

// HasErrors retorna true se há erros acumulados
func (b *BackupBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *BackupBuilder) Errors() []error {
	return b.errs
}

