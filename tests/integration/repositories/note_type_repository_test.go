package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

func TestNoteTypeRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_save")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Basic").
		WithFieldsJSON(`[{"name":"Front"},{"name":"Back"}]`).
		WithCardTypesJSON(`[{"name":"Card 1"}]`).
		WithTemplatesJSON(`[{"name":"Template 1"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)
	assert.Greater(t, noteTypeEntity.GetID(), int64(0))
}

func TestNoteTypeRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_find")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Cloze").
		WithFieldsJSON(`[{"name":"Text"}]`).
		WithCardTypesJSON(`[{"name":"Cloze"}]`).
		WithTemplatesJSON(`[{"name":"Cloze Template"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)
	noteTypeID := noteTypeEntity.GetID()

	found, err := noteTypeRepo.FindByID(ctx, userID, noteTypeID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, noteTypeID, found.GetID())
	assert.Equal(t, "Cloze", found.GetName())
}

func TestNoteTypeRepository_FindByName(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_name")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("Unique Note Type").
		WithFieldsJSON(`[{"name":"Field"}]`).
		WithCardTypesJSON(`[{"name":"Card"}]`).
		WithTemplatesJSON(`[{"name":"Template"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)

	found, err := noteTypeRepo.FindByName(ctx, userID, "Unique Note Type")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Unique Note Type", found.GetName())
}

func TestNoteTypeRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	noteTypeRepo := repositories.NewNoteTypeRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "notetype_delete")

	noteTypeEntity, err := notetype.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithName("To Delete").
		WithFieldsJSON(`[{"name":"Field"}]`).
		WithCardTypesJSON(`[{"name":"Card"}]`).
		WithTemplatesJSON(`[{"name":"Template"}]`).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = noteTypeRepo.Save(ctx, userID, noteTypeEntity)
	require.NoError(t, err)
	noteTypeID := noteTypeEntity.GetID()

	err = noteTypeRepo.Delete(ctx, userID, noteTypeID)
	require.NoError(t, err)

	found, err := noteTypeRepo.FindByID(ctx, userID, noteTypeID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ownership.ErrResourceNotFound)
	assert.Nil(t, found)
}

