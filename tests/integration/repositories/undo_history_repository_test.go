package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestUndoHistoryRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	undoHistoryRepo := repositories.NewUndoHistoryRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "undo_history_save")

	undoHistoryEntity, err := undohistory.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithOperationType("delete_note").
		WithOperationData(`{"note_id":123}`).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = undoHistoryRepo.Save(ctx, userID, undoHistoryEntity)
	require.NoError(t, err)
	assert.Greater(t, undoHistoryEntity.GetID(), int64(0))
}

func TestUndoHistoryRepository_FindLatest(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	undoHistoryRepo := repositories.NewUndoHistoryRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "undo_history_latest")

	undoHistoryEntity, err := undohistory.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithOperationType("move_card").
		WithOperationData(`{"card_id":789}`).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = undoHistoryRepo.Save(ctx, userID, undoHistoryEntity)
	require.NoError(t, err)

	found, err := undoHistoryRepo.FindLatest(ctx, userID, 1)
	require.NoError(t, err)
	assert.Greater(t, len(found), 0)
}

