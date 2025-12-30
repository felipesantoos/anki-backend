package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/felipesantos/anki-backend/core/domain/entities/user"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestUserRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repositories.NewUserRepository(db.DB)

	email, _ := valueobjects.NewEmail("test_save@example.com")
	password, _ := valueobjects.NewPassword("password123")
	now := time.Now()

	userEntity, err := user.NewBuilder().
		WithID(0).
		WithEmail(email).
		WithPasswordHash(password).
		WithEmailVerified(false).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	require.NoError(t, err)

	err = repo.Save(ctx, userEntity)
	require.NoError(t, err)
	assert.Greater(t, userEntity.GetID(), int64(0))

	// Verify it was saved
	found, err := repo.FindByID(ctx, userEntity.GetID())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, email.Value(), found.GetEmail().Value())
}

func TestUserRepository_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repositories.NewUserRepository(db.DB)

	userID, _ := createTestUser(t, ctx, repo, "find_by_id")

	found, err := repo.FindByID(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userID, found.GetID())
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repositories.NewUserRepository(db.DB)

	emailStr := "test_find_email@example.com"
	email, _ := valueobjects.NewEmail(emailStr)
	password, _ := valueobjects.NewPassword("password123")
	now := time.Now()

	userEntity, _ := user.NewBuilder().
		WithID(0).
		WithEmail(email).
		WithPasswordHash(password).
		WithEmailVerified(false).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	
	_ = repo.Save(ctx, userEntity)

	found, err := repo.FindByEmail(ctx, emailStr)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userEntity.GetID(), found.GetID())
}

func TestUserRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repositories.NewUserRepository(db.DB)

	userID, _ := createTestUser(t, ctx, repo, "update")
	userEntity, _ := repo.FindByID(ctx, userID)

	userEntity.SetEmailVerified(true)
	now := time.Now()
	userEntity.SetLastLoginAt(&now)

	err := repo.Update(ctx, userEntity)
	require.NoError(t, err)

	updated, err := repo.FindByID(ctx, userID)
	require.NoError(t, err)
	assert.True(t, updated.GetEmailVerified())
	assert.NotNil(t, updated.GetLastLoginAt())
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repositories.NewUserRepository(db.DB)

	_, email := createTestUser(t, ctx, repo, "exists")

	exists, err := repo.ExistsByEmail(ctx, email)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByEmail(ctx, "nonexistent@example.com")
	require.NoError(t, err)
	assert.False(t, exists)
}

