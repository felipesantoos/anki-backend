package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestBackupRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	backupRepo := repositories.NewBackupRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "backup_save")

	backupEntity, err := backup.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("backup_20240101.apkg").
		WithSize(5000000).
		WithStoragePath("/storage/backup_20240101.apkg").
		WithBackupType("automatic").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = backupRepo.Save(ctx, userID, backupEntity)
	require.NoError(t, err)
	assert.Greater(t, backupEntity.GetID(), int64(0))
}

func TestBackupRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	backupRepo := repositories.NewBackupRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "backup_find")

	backupEntity, err := backup.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("test_backup.apkg").
		WithSize(1000000).
		WithStoragePath("/storage/test_backup.apkg").
		WithBackupType("manual").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = backupRepo.Save(ctx, userID, backupEntity)
	require.NoError(t, err)
	backupID := backupEntity.GetID()

	found, err := backupRepo.FindByID(ctx, userID, backupID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, backupID, found.GetID())
}

func TestBackupRepository_FindByFilename(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	backupRepo := repositories.NewBackupRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "backup_filename")

	backupEntity, err := backup.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("unique_backup.apkg").
		WithSize(2000000).
		WithStoragePath("/storage/unique_backup.apkg").
		WithBackupType("automatic").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = backupRepo.Save(ctx, userID, backupEntity)
	require.NoError(t, err)

	found, err := backupRepo.FindByFilename(ctx, userID, "unique_backup.apkg")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "unique_backup.apkg", found.GetFilename())
}

func TestBackupRepository_FindByType(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	backupRepo := repositories.NewBackupRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "backup_type")

	// Create automatic backup
	backup1, err := backup.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("auto1.apkg").
		WithSize(1000000).
		WithStoragePath("/storage/auto1.apkg").
		WithBackupType("automatic").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = backupRepo.Save(ctx, userID, backup1)
	require.NoError(t, err)

	// Create manual backup
	backup2, err := backup.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithFilename("manual1.apkg").
		WithSize(2000000).
		WithStoragePath("/storage/manual1.apkg").
		WithBackupType("manual").
		WithCreatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	err = backupRepo.Save(ctx, userID, backup2)
	require.NoError(t, err)

	// Find automatic backups
	automaticBackups, err := backupRepo.FindByType(ctx, userID, "automatic")
	require.NoError(t, err)
	assert.Len(t, automaticBackups, 1)
	assert.Equal(t, "automatic", automaticBackups[0].GetBackupType())
}

