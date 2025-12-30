package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	"github.com/stretchr/testify/assert"
)

func TestToSyncMetaResponse(t *testing.T) {
	now := time.Now()
	
	sm, _ := syncmeta.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithClientID("client123").
		WithLastSync(now).
		WithLastSyncUSN(100).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToSyncMetaResponse(sm)
		assert.NotNil(t, res)
		assert.Equal(t, sm.GetID(), res.ID)
		assert.Equal(t, sm.GetUserID(), res.UserID)
		assert.Equal(t, sm.GetClientID(), res.ClientID)
		assert.Equal(t, sm.GetLastSync(), res.LastSync)
		assert.Equal(t, sm.GetLastSyncUSN(), res.LastSyncUSN)
		assert.Equal(t, sm.GetCreatedAt(), res.CreatedAt)
		assert.Equal(t, sm.GetUpdatedAt(), res.UpdatedAt)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToSyncMetaResponse(nil)
		assert.Nil(t, res)
	})
}
