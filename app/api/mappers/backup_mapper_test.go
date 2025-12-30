package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/stretchr/testify/assert"
)

func TestToBackupResponse(t *testing.T) {
	now := time.Now()
	
	b, _ := backup.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithFilename("backup.colpkg").
		WithSize(1024).
		WithStoragePath("/path/to/backup").
		WithBackupType(backup.BackupTypeAutomatic).
		WithCreatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToBackupResponse(b)
		assert.NotNil(t, res)
		assert.Equal(t, b.GetID(), res.ID)
		assert.Equal(t, b.GetUserID(), res.UserID)
		assert.Equal(t, b.GetFilename(), res.Filename)
		assert.Equal(t, b.GetSize(), res.Size)
		assert.Equal(t, b.GetStoragePath(), res.StoragePath)
		assert.Equal(t, b.GetBackupType(), res.BackupType)
		assert.Equal(t, b.GetCreatedAt(), res.CreatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToBackupResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToBackupResponseList(t *testing.T) {
	b1, _ := backup.NewBuilder().WithID(1).Build()
	b2, _ := backup.NewBuilder().WithID(2).Build()
	backups := []*backup.Backup{b1, b2}

	res := ToBackupResponseList(backups)
	assert.Len(t, res, 2)
	assert.Equal(t, b1.GetID(), res[0].ID)
	assert.Equal(t, b2.GetID(), res[1].ID)
}
