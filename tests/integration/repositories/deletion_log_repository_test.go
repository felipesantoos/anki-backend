package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestDeletionLogRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deletionLogRepo := repositories.NewDeletionLogRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deletion_log_save")

	deletionLogEntity, err := deletionlog.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithObjectType("note").
		WithObjectID(123).
		WithObjectData(`{"id":123,"fields":{"Front":"Test"}}`).
		WithDeletedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = deletionLogRepo.Save(ctx, userID, deletionLogEntity)
	require.NoError(t, err)
	assert.Greater(t, deletionLogEntity.GetID(), int64(0))
}

func TestDeletionLogRepository_FindByObjectType(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deletionLogRepo := repositories.NewDeletionLogRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "deletion_log_type")

	deletionLogEntity, err := deletionlog.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithObjectType("card").
		WithObjectID(456).
		WithObjectData(`{"id":456}`).
		WithDeletedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = deletionLogRepo.Save(ctx, userID, deletionLogEntity)
	require.NoError(t, err)

	found, err := deletionLogRepo.FindByObjectType(ctx, userID, "card")
	require.NoError(t, err)
	assert.Greater(t, len(found), 0)
}

