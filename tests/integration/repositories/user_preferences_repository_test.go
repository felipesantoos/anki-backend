package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/repositories"
)

func TestUserPreferencesRepository_Save_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	prefsRepo := repositories.NewUserPreferencesRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "prefs_save")

	nextDayTime := time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)
	prefsEntity, err := userpreferences.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithLanguage("pt-BR").
		WithTheme(valueobjects.ThemeTypeLight).
		WithAutoSync(true).
		WithNextDayStartsAt(nextDayTime).
		WithLearnAheadLimit(20).
		WithTimeboxTimeLimit(30).
		WithVideoDriver("opengl").
		WithUISize(1.0).
		WithMinimalistMode(false).
		WithReduceMotion(false).
		WithPasteStripsFormatting(true).
		WithPasteImagesAsPNG(false).
		WithDefaultDeckBehavior("add").
		WithShowPlayButtons(true).
		WithInterruptAudioOnAnswer(false).
		WithShowRemainingCount(true).
		WithShowNextReviewTime(true).
		WithSpacebarAnswersCard(true).
		WithIgnoreAccentsInSearch(false).
		WithSyncAudioAndImages(true).
		WithPeriodicallySyncMedia(false).
		WithForceOneWaySync(false).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = prefsRepo.Save(ctx, userID, prefsEntity)
	require.NoError(t, err)
	assert.Greater(t, prefsEntity.GetID(), int64(0))
}

func TestUserPreferencesRepository_FindByUserID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userRepo := repositories.NewUserRepository(db.DB)
	prefsRepo := repositories.NewUserPreferencesRepository(db.DB)

	userID, _ := createTestUser(t, ctx, userRepo, "prefs_find")

	nextDayTime := time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)
	prefsEntity, err := userpreferences.NewBuilder().
		WithID(0).
		WithUserID(userID).
		WithLanguage("en-US").
		WithTheme(valueobjects.ThemeTypeDark).
		WithAutoSync(false).
		WithNextDayStartsAt(nextDayTime).
		WithLearnAheadLimit(10).
		WithTimeboxTimeLimit(15).
		WithVideoDriver("software").
		WithUISize(0.8).
		WithMinimalistMode(true).
		WithReduceMotion(true).
		WithPasteStripsFormatting(false).
		WithPasteImagesAsPNG(true).
		WithDefaultDeckBehavior("create").
		WithShowPlayButtons(false).
		WithInterruptAudioOnAnswer(true).
		WithShowRemainingCount(false).
		WithShowNextReviewTime(false).
		WithSpacebarAnswersCard(false).
		WithIgnoreAccentsInSearch(true).
		WithSyncAudioAndImages(false).
		WithPeriodicallySyncMedia(true).
		WithForceOneWaySync(true).
		WithCreatedAt(time.Now()).
		WithUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	err = prefsRepo.Save(ctx, userID, prefsEntity)
	require.NoError(t, err)

	found, err := prefsRepo.FindByUserID(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userID, found.GetUserID())
	assert.Equal(t, "en-US", found.GetLanguage())
}

