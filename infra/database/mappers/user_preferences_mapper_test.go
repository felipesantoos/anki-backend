package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestUserPreferencesToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	nextDayStartsAt := time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC)
	defaultSearchText := "default search"
	selfHostedURL := "https://sync.example.com"

	model := &models.UserPreferencesModel{
		ID:                        1,
		UserID:                    100,
		Language:                  "pt-BR",
		Theme:                     valueobjects.ThemeTypeLight.String(),
		AutoSync:                  true,
		NextDayStartsAt:           nextDayStartsAt,
		LearnAheadLimit:           20,
		TimeboxTimeLimit:          30,
		VideoDriver:               "opengl",
		UISize:                    1.0,
		MinimalistMode:            false,
		ReduceMotion:              false,
		PasteStripsFormatting:     true,
		PasteImagesAsPNG:          false,
		DefaultDeckBehavior:       "add",
		ShowPlayButtons:           true,
		InterruptAudioOnAnswer:    false,
		ShowRemainingCount:        true,
		ShowNextReviewTime:        true,
		SpacebarAnswersCard:       true,
		IgnoreAccentsInSearch:     false,
		DefaultSearchText:         sqlNullString(defaultSearchText, true),
		SyncAudioAndImages:        true,
		PeriodicallySyncMedia:     false,
		ForceOneWaySync:           false,
		SelfHostedSyncServerURL:   sqlNullString(selfHostedURL, true),
		CreatedAt:                 now,
		UpdatedAt:                 now.Add(time.Hour),
	}

	entity, err := UserPreferencesToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "pt-BR", entity.GetLanguage())
	assert.Equal(t, valueobjects.ThemeTypeLight, entity.GetTheme())
	assert.True(t, entity.GetAutoSync())
	// NextDayStartsAt is stored as TIME, so only time portion is preserved
	// The mapper normalizes it to 1970-01-01, but we only care about the time
	expectedTime := time.Date(1970, 1, 1, nextDayStartsAt.Hour(), nextDayStartsAt.Minute(), nextDayStartsAt.Second(), 0, time.UTC)
	assert.Equal(t, expectedTime, entity.GetNextDayStartsAt())
	assert.Equal(t, 20, entity.GetLearnAheadLimit())
	assert.Equal(t, 30, entity.GetTimeboxTimeLimit())
	assert.Equal(t, "opengl", entity.GetVideoDriver())
	assert.Equal(t, 1.0, entity.GetUISize())
	assert.False(t, entity.GetMinimalistMode())
	assert.False(t, entity.GetReduceMotion())
	assert.True(t, entity.GetPasteStripsFormatting())
	assert.False(t, entity.GetPasteImagesAsPNG())
	assert.Equal(t, "add", entity.GetDefaultDeckBehavior())
	assert.True(t, entity.GetShowPlayButtons())
	assert.False(t, entity.GetInterruptAudioOnAnswer())
	assert.True(t, entity.GetShowRemainingCount())
	assert.True(t, entity.GetShowNextReviewTime())
	assert.True(t, entity.GetSpacebarAnswersCard())
	assert.False(t, entity.GetIgnoreAccentsInSearch())
	assert.NotNil(t, entity.GetDefaultSearchText())
	assert.Equal(t, defaultSearchText, *entity.GetDefaultSearchText())
	assert.True(t, entity.GetSyncAudioAndImages())
	assert.False(t, entity.GetPeriodicallySyncMedia())
	assert.False(t, entity.GetForceOneWaySync())
	assert.NotNil(t, entity.GetSelfHostedSyncServerURL())
	assert.Equal(t, selfHostedURL, *entity.GetSelfHostedSyncServerURL())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
}

func TestUserPreferencesToDomain_WithNullFields(t *testing.T) {
	now := time.Now()
	nextDayStartsAt := time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC)

	model := &models.UserPreferencesModel{
		ID:                        2,
		UserID:                    200,
		Language:                  "en-US",
		Theme:                     valueobjects.ThemeTypeDark.String(),
		AutoSync:                  false,
		NextDayStartsAt:           nextDayStartsAt,
		LearnAheadLimit:           10,
		TimeboxTimeLimit:          15,
		VideoDriver:               "software",
		UISize:                    0.8,
		MinimalistMode:            true,
		ReduceMotion:              true,
		PasteStripsFormatting:     false,
		PasteImagesAsPNG:          true,
		DefaultDeckBehavior:       "create",
		ShowPlayButtons:           false,
		InterruptAudioOnAnswer:    true,
		ShowRemainingCount:        false,
		ShowNextReviewTime:        false,
		SpacebarAnswersCard:       false,
		IgnoreAccentsInSearch:     true,
		DefaultSearchText:         sqlNullString("", false),
		SyncAudioAndImages:        false,
		PeriodicallySyncMedia:     true,
		ForceOneWaySync:           true,
		SelfHostedSyncServerURL:   sqlNullString("", false),
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}

	entity, err := UserPreferencesToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetDefaultSearchText())
	assert.Nil(t, entity.GetSelfHostedSyncServerURL())
}

func TestUserPreferencesToDomain_NilInput(t *testing.T) {
	entity, err := UserPreferencesToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestUserPreferencesToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	nextDayStartsAt := time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC)
	defaultSearchText := "default search"
	selfHostedURL := "https://sync.example.com"

	entity, err := userpreferences.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithLanguage("pt-BR").
		WithTheme(valueobjects.ThemeTypeLight).
		WithAutoSync(true).
		WithNextDayStartsAt(nextDayStartsAt).
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
		WithDefaultSearchText(&defaultSearchText).
		WithSyncAudioAndImages(true).
		WithPeriodicallySyncMedia(false).
		WithForceOneWaySync(false).
		WithSelfHostedSyncServerURL(&selfHostedURL).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		Build()
	require.NoError(t, err)

	model := UserPreferencesToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "pt-BR", model.Language)
	assert.Equal(t, valueobjects.ThemeTypeLight.String(), model.Theme)
	assert.True(t, model.AutoSync)
	// NextDayStartsAt is stored as TIME, so date is normalized to 1970-01-01
	expectedTime := time.Date(1970, 1, 1, nextDayStartsAt.Hour(), nextDayStartsAt.Minute(), nextDayStartsAt.Second(), 0, time.UTC)
	assert.Equal(t, expectedTime, model.NextDayStartsAt)
	assert.Equal(t, 20, model.LearnAheadLimit)
	assert.Equal(t, 30, model.TimeboxTimeLimit)
	assert.Equal(t, "opengl", model.VideoDriver)
	assert.Equal(t, 1.0, model.UISize)
	assert.False(t, model.MinimalistMode)
	assert.False(t, model.ReduceMotion)
	assert.True(t, model.PasteStripsFormatting)
	assert.False(t, model.PasteImagesAsPNG)
	assert.Equal(t, "add", model.DefaultDeckBehavior)
	assert.True(t, model.ShowPlayButtons)
	assert.False(t, model.InterruptAudioOnAnswer)
	assert.True(t, model.ShowRemainingCount)
	assert.True(t, model.ShowNextReviewTime)
	assert.True(t, model.SpacebarAnswersCard)
	assert.False(t, model.IgnoreAccentsInSearch)
	assert.True(t, model.DefaultSearchText.Valid)
	assert.Equal(t, defaultSearchText, model.DefaultSearchText.String)
	assert.True(t, model.SyncAudioAndImages)
	assert.False(t, model.PeriodicallySyncMedia)
	assert.False(t, model.ForceOneWaySync)
	assert.True(t, model.SelfHostedSyncServerURL.Valid)
	assert.Equal(t, selfHostedURL, model.SelfHostedSyncServerURL.String)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
}

func TestUserPreferencesToModel_WithNullFields(t *testing.T) {
	now := time.Now()
	nextDayStartsAt := time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC)

	entity, err := userpreferences.NewBuilder().
		WithID(2).
		WithUserID(200).
		WithLanguage("en-US").
		WithTheme(valueobjects.ThemeTypeDark).
		WithAutoSync(false).
		WithNextDayStartsAt(nextDayStartsAt).
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
		WithDefaultSearchText(nil).
		WithSyncAudioAndImages(false).
		WithPeriodicallySyncMedia(true).
		WithForceOneWaySync(true).
		WithSelfHostedSyncServerURL(nil).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	require.NoError(t, err)

	model := UserPreferencesToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.DefaultSearchText.Valid)
	assert.False(t, model.SelfHostedSyncServerURL.Valid)
}

func TestUserPreferencesToDomain_AllThemes(t *testing.T) {
	themes := []valueobjects.ThemeType{
		valueobjects.ThemeTypeLight,
		valueobjects.ThemeTypeDark,
		valueobjects.ThemeTypeAuto,
	}

	for _, theme := range themes {
		t.Run(theme.String(), func(t *testing.T) {
			nextDayStartsAt := time.Date(2024, 1, 1, 4, 0, 0, 0, time.UTC)
			model := &models.UserPreferencesModel{
				ID:               1,
				UserID:           100,
				Language:         "en-US",
				Theme:            theme.String(),
				AutoSync:         true,
				NextDayStartsAt:  nextDayStartsAt,
				LearnAheadLimit: 20,
				TimeboxTimeLimit: 30,
				VideoDriver:      "opengl",
				UISize:           1.0,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			entity, err := UserPreferencesToDomain(model)
			require.NoError(t, err)
			assert.Equal(t, theme, entity.GetTheme())
		})
	}
}

