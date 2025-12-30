package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/stretchr/testify/assert"
)

func TestToMediaResponse(t *testing.T) {
	now := time.Now()
	
	m, _ := media.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithFilename("image.png").
		WithHash("hash123").
		WithSize(512).
		WithMimeType("image/png").
		WithStoragePath("/path/to/media").
		WithCreatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToMediaResponse(m)
		assert.NotNil(t, res)
		assert.Equal(t, m.GetID(), res.ID)
		assert.Equal(t, m.GetUserID(), res.UserID)
		assert.Equal(t, m.GetFilename(), res.Filename)
		assert.Equal(t, m.GetHash(), res.Hash)
		assert.Equal(t, m.GetSize(), res.Size)
		assert.Equal(t, m.GetMimeType(), res.MimeType)
		assert.Equal(t, m.GetStoragePath(), res.StoragePath)
		assert.Equal(t, m.GetCreatedAt(), res.CreatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToMediaResponse(nil)
		assert.Nil(t, res)
	})
}

func TestToMediaResponseList(t *testing.T) {
	m1, _ := media.NewBuilder().WithID(1).WithUserID(1).WithFilename("f1").Build()
	m2, _ := media.NewBuilder().WithID(2).WithUserID(1).WithFilename("f2").Build()
	mediaFiles := []*media.Media{m1, m2}

	res := ToMediaResponseList(mediaFiles)
	assert.Len(t, res, 2)
	assert.Equal(t, m1.GetID(), res[0].ID)
	assert.Equal(t, m2.GetID(), res[1].ID)
}
