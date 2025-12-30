package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestSyncMetaToDomain_WithAllFields(t *testing.T) {
	now := time.Now()

	model := &models.SyncMetaModel{
		ID:          1,
		UserID:      100,
		ClientID:    "client-123",
		LastSync:    now,
		LastSyncUSN: 42,
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Hour),
	}

	entity, err := SyncMetaToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "client-123", entity.GetClientID())
	assert.Equal(t, now, entity.GetLastSync())
	assert.Equal(t, int64(42), entity.GetLastSyncUSN())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
}

func TestSyncMetaToDomain_NilInput(t *testing.T) {
	entity, err := SyncMetaToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestSyncMetaToModel_WithAllFields(t *testing.T) {
	now := time.Now()

	entity, err := syncmeta.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithClientID("client-123").
		WithLastSync(now).
		WithLastSyncUSN(42).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		Build()
	require.NoError(t, err)

	model := SyncMetaToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "client-123", model.ClientID)
	assert.Equal(t, now, model.LastSync)
	assert.Equal(t, int64(42), model.LastSyncUSN)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
}

