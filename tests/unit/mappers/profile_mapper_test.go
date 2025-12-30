package mappers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestProfileToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	ankiWebUsername := "testuser"

	model := &models.ProfileModel{
		ID:                1,
		UserID:            100,
		Name:              "Default Profile",
		AnkiWebSyncEnabled: true,
		AnkiWebUsername:    sqlNullString(ankiWebUsername, true),
		CreatedAt:         now,
		UpdatedAt:         now.Add(time.Hour),
		DeletedAt:         sqlNullTime(deletedAt, true),
	}

	entity, err := ProfileToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "Default Profile", entity.GetName())
	assert.True(t, entity.GetAnkiWebSyncEnabled())
	assert.NotNil(t, entity.GetAnkiWebUsername())
	assert.Equal(t, ankiWebUsername, *entity.GetAnkiWebUsername())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
	assert.NotNil(t, entity.GetDeletedAt())
	assert.Equal(t, deletedAt, *entity.GetDeletedAt())
}

func TestProfileToDomain_WithNullFields(t *testing.T) {
	now := time.Now()

	model := &models.ProfileModel{
		ID:                2,
		UserID:            200,
		Name:              "Profile 2",
		AnkiWebSyncEnabled: false,
		AnkiWebUsername:    sqlNullString("", false),
		CreatedAt:         now,
		UpdatedAt:         now,
		DeletedAt:         sqlNullTime(time.Time{}, false),
	}

	entity, err := ProfileToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetAnkiWebUsername())
	assert.Nil(t, entity.GetDeletedAt())
}

func TestProfileToDomain_NilInput(t *testing.T) {
	entity, err := ProfileToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestProfileToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)
	ankiWebUsername := "testuser"

	entity, err := profile.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithName("Default Profile").
		WithAnkiWebSyncEnabled(true).
		WithAnkiWebUsername(&ankiWebUsername).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		WithDeletedAt(&deletedAt).
		Build()
	require.NoError(t, err)

	model := ProfileToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "Default Profile", model.Name)
	assert.True(t, model.AnkiWebSyncEnabled)
	assert.True(t, model.AnkiWebUsername.Valid)
	assert.Equal(t, ankiWebUsername, model.AnkiWebUsername.String)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestProfileToModel_WithNullFields(t *testing.T) {
	now := time.Now()

	entity, err := profile.NewBuilder().
		WithID(2).
		WithUserID(200).
		WithName("Profile 2").
		WithAnkiWebSyncEnabled(false).
		WithAnkiWebUsername(nil).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		WithDeletedAt(nil).
		Build()
	require.NoError(t, err)

	model := ProfileToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.AnkiWebUsername.Valid)
	assert.False(t, model.DeletedAt.Valid)
}

