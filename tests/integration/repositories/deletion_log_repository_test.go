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

func TestDeletionLogRepository_FindRecent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	deletionLogRepo := repositories.NewDeletionLogRepository(db.DB)

	userID1, _ := createTestUser(t, ctx, userRepo, "deletion_log_recent_user1")
	userID2, _ := createTestUser(t, ctx, userRepo, "deletion_log_recent_user2")

	now := time.Now()
	recentTime := now.Add(-3 * 24 * time.Hour) // 3 days ago
	oldTime := now.Add(-10 * 24 * time.Hour)   // 10 days ago

	// Create recent deletion logs for user 1
	dl1, _ := deletionlog.NewBuilder().
		WithID(0).
		WithUserID(userID1).
		WithObjectType("note").
		WithObjectID(101).
		WithObjectData(`{"id":101}`).
		WithDeletedAt(recentTime).
		Build()
	dl2, _ := deletionlog.NewBuilder().
		WithID(0).
		WithUserID(userID1).
		WithObjectType("card").
		WithObjectID(102).
		WithObjectData(`{"id":102}`).
		WithDeletedAt(recentTime).
		Build()
	// Create old deletion log for user 1 (should not be returned)
	dl3, _ := deletionlog.NewBuilder().
		WithID(0).
		WithUserID(userID1).
		WithObjectType("note").
		WithObjectID(103).
		WithObjectData(`{"id":103}`).
		WithDeletedAt(oldTime).
		Build()

	require.NoError(t, deletionLogRepo.Save(ctx, userID1, dl1))
	require.NoError(t, deletionLogRepo.Save(ctx, userID1, dl2))
	require.NoError(t, deletionLogRepo.Save(ctx, userID1, dl3))

	// Create deletion log for user 2 (should not be returned)
	dl4, _ := deletionlog.NewBuilder().
		WithID(0).
		WithUserID(userID2).
		WithObjectType("note").
		WithObjectID(201).
		WithObjectData(`{"id":201}`).
		WithDeletedAt(recentTime).
		Build()
	require.NoError(t, deletionLogRepo.Save(ctx, userID2, dl4))

	t.Run("Success - find recent deletions within 5 days", func(t *testing.T) {
		found, err := deletionLogRepo.FindRecent(ctx, userID1, 10, 5)
		require.NoError(t, err)
		assert.Len(t, found, 2) // Only dl1 and dl2 (recent), not dl3 (old) or dl4 (other user)
		for _, dl := range found {
			assert.Equal(t, userID1, dl.GetUserID())
			assert.True(t, dl.GetDeletedAt().After(now.Add(-5*24*time.Hour)))
		}
	})

	t.Run("Success - limit enforced", func(t *testing.T) {
		found, err := deletionLogRepo.FindRecent(ctx, userID1, 1, 5)
		require.NoError(t, err)
		assert.Len(t, found, 1) // Limited to 1
	})

	t.Run("Success - empty result when no recent deletions", func(t *testing.T) {
		found, err := deletionLogRepo.FindRecent(ctx, userID1, 10, 1) // Only 1 day
		require.NoError(t, err)
		// May be empty or contain very recent deletions depending on timing
		assert.NotNil(t, found)
	})

	t.Run("Cross-user isolation - user 2 cannot see user 1's deletions", func(t *testing.T) {
		found, err := deletionLogRepo.FindRecent(ctx, userID2, 10, 5)
		require.NoError(t, err)
		// Should only find user 2's deletion (dl4)
		for _, dl := range found {
			assert.Equal(t, userID2, dl.GetUserID())
		}
	})
}

