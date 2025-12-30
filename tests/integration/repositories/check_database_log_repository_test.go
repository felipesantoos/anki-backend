package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	checkdatabaselog "github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestCheckDatabaseLogRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	checkLogRepo := repositories.NewCheckDatabaseLogRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "check_log_save")

	execTimeMs := 1500
	checkLogEntity, err := checkdatabaselog.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithStatus(checkdatabaselog.CheckStatusCompleted).
		WithIssuesFound(0).
		WithIssuesDetails(`[]`).
		WithExecutionTimeMs(&execTimeMs).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = checkLogRepo.Save(ctx, userID, checkLogEntity)
	require.NoError(t, err)
	assert.Greater(t, checkLogEntity.GetID(), int64(0))
}

func TestCheckDatabaseLogRepository_FindLatest(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	checkLogRepo := repositories.NewCheckDatabaseLogRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "check_log_latest")

	checkLogEntity, err := checkdatabaselog.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithStatus(checkdatabaselog.CheckStatusFailed).
		WithIssuesFound(5).
		WithIssuesDetails(`[{"issue":"missing note"}]`).
		WithExecutionTimeMs(nil).
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = checkLogRepo.Save(ctx, userID, checkLogEntity)
	require.NoError(t, err)

	found, err := checkLogRepo.FindLatest(ctx, userID, 1)
	require.NoError(t, err)
	assert.Greater(t, len(found), 0)
}

