package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestDeletionLogToDomain_WithAllFields(t *testing.T) {
	now := time.Now()

	model := &models.DeletionLogModel{
		ID:         1,
		UserID:     100,
		ObjectType: "note",
		ObjectID:   123,
		ObjectData: `{"id":123,"guid":"test-guid"}`,
		DeletedAt:  now,
	}

	entity, err := DeletionLogToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "note", entity.GetObjectType())
	assert.Equal(t, int64(123), entity.GetObjectID())
	assert.Equal(t, `{"id":123,"guid":"test-guid"}`, entity.GetObjectData())
	assert.Equal(t, now, entity.GetDeletedAt())
}

func TestDeletionLogToDomain_NilInput(t *testing.T) {
	entity, err := DeletionLogToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestDeletionLogToModel_WithAllFields(t *testing.T) {
	now := time.Now()

	entity, err := deletionlog.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithObjectType("note").
		WithObjectID(123).
		WithObjectData(`{"id":123}`).
		WithDeletedAt(now).
		Build()
	require.NoError(t, err)

	model := DeletionLogToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "note", model.ObjectType)
	assert.Equal(t, int64(123), model.ObjectID)
}

