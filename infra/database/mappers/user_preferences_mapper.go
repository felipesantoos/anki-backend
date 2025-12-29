package mappers

import (
	"database/sql"
	"fmt"
	"time"

	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// UserPreferencesToDomain converts a UserPreferencesModel (database representation) to a UserPreferences entity (domain representation)
func UserPreferencesToDomain(model *models.UserPreferencesModel) (*userpreferences.UserPreferences, error) {
	if model == nil {
		return nil, nil
	}

	// Parse theme from string
	theme := valueobjects.ThemeType(model.Theme)
	if !theme.IsValid() {
		return nil, fmt.Errorf("invalid theme type: %s", model.Theme)
	}

	// For next_day_starts_at (TIME), extract just the time portion
	// PostgreSQL TIME is stored as time.Time but we only need the time component
	nextDayTime := time.Date(1970, 1, 1, model.NextDayStartsAt.Hour(), model.NextDayStartsAt.Minute(), model.NextDayStartsAt.Second(), 0, time.UTC)

	builder := userpreferences.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithLanguage(model.Language).
		WithTheme(theme).
		WithAutoSync(model.AutoSync).
		WithNextDayStartsAt(nextDayTime).
		WithLearnAheadLimit(model.LearnAheadLimit).
		WithTimeboxTimeLimit(model.TimeboxTimeLimit).
		WithVideoDriver(model.VideoDriver).
		WithUISize(model.UISize).
		WithMinimalistMode(model.MinimalistMode).
		WithReduceMotion(model.ReduceMotion).
		WithPasteStripsFormatting(model.PasteStripsFormatting).
		WithPasteImagesAsPNG(model.PasteImagesAsPNG).
		WithDefaultDeckBehavior(model.DefaultDeckBehavior).
		WithShowPlayButtons(model.ShowPlayButtons).
		WithInterruptAudioOnAnswer(model.InterruptAudioOnAnswer).
		WithShowRemainingCount(model.ShowRemainingCount).
		WithShowNextReviewTime(model.ShowNextReviewTime).
		WithSpacebarAnswersCard(model.SpacebarAnswersCard).
		WithIgnoreAccentsInSearch(model.IgnoreAccentsInSearch).
		WithSyncAudioAndImages(model.SyncAudioAndImages).
		WithPeriodicallySyncMedia(model.PeriodicallySyncMedia).
		WithForceOneWaySync(model.ForceOneWaySync).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable default_search_text
	if model.DefaultSearchText.Valid {
		builder = builder.WithDefaultSearchText(&model.DefaultSearchText.String)
	}

	// Handle nullable self_hosted_sync_server_url
	if model.SelfHostedSyncServerURL.Valid {
		builder = builder.WithSelfHostedSyncServerURL(&model.SelfHostedSyncServerURL.String)
	}

	return builder.Build()
}

// UserPreferencesToModel converts a UserPreferences entity (domain representation) to a UserPreferencesModel (database representation)
func UserPreferencesToModel(prefsEntity *userpreferences.UserPreferences) *models.UserPreferencesModel {
	// For next_day_starts_at, we need to store it as a TIME value
	// Use a fixed date (1970-01-01) with the time from the entity
	nextDayTime := time.Date(1970, 1, 1,
		prefsEntity.GetNextDayStartsAt().Hour(),
		prefsEntity.GetNextDayStartsAt().Minute(),
		prefsEntity.GetNextDayStartsAt().Second(),
		0, time.UTC)

	model := &models.UserPreferencesModel{
		ID:                      prefsEntity.GetID(),
		UserID:                  prefsEntity.GetUserID(),
		Language:                prefsEntity.GetLanguage(),
		Theme:                   prefsEntity.GetTheme().String(),
		AutoSync:                prefsEntity.GetAutoSync(),
		NextDayStartsAt:         nextDayTime,
		LearnAheadLimit:         prefsEntity.GetLearnAheadLimit(),
		TimeboxTimeLimit:        prefsEntity.GetTimeboxTimeLimit(),
		VideoDriver:             prefsEntity.GetVideoDriver(),
		UISize:                  prefsEntity.GetUISize(),
		MinimalistMode:          prefsEntity.GetMinimalistMode(),
		ReduceMotion:            prefsEntity.GetReduceMotion(),
		PasteStripsFormatting:   prefsEntity.GetPasteStripsFormatting(),
		PasteImagesAsPNG:        prefsEntity.GetPasteImagesAsPNG(),
		DefaultDeckBehavior:     prefsEntity.GetDefaultDeckBehavior(),
		ShowPlayButtons:         prefsEntity.GetShowPlayButtons(),
		InterruptAudioOnAnswer:  prefsEntity.GetInterruptAudioOnAnswer(),
		ShowRemainingCount:      prefsEntity.GetShowRemainingCount(),
		ShowNextReviewTime:      prefsEntity.GetShowNextReviewTime(),
		SpacebarAnswersCard:     prefsEntity.GetSpacebarAnswersCard(),
		IgnoreAccentsInSearch:   prefsEntity.GetIgnoreAccentsInSearch(),
		SyncAudioAndImages:      prefsEntity.GetSyncAudioAndImages(),
		PeriodicallySyncMedia:   prefsEntity.GetPeriodicallySyncMedia(),
		ForceOneWaySync:         prefsEntity.GetForceOneWaySync(),
		CreatedAt:               prefsEntity.GetCreatedAt(),
		UpdatedAt:               prefsEntity.GetUpdatedAt(),
	}

	// Handle nullable default_search_text
	if prefsEntity.GetDefaultSearchText() != nil {
		model.DefaultSearchText = sql.NullString{
			String: *prefsEntity.GetDefaultSearchText(),
			Valid:  true,
		}
	}

	// Handle nullable self_hosted_sync_server_url
	if prefsEntity.GetSelfHostedSyncServerURL() != nil {
		model.SelfHostedSyncServerURL = sql.NullString{
			String: *prefsEntity.GetSelfHostedSyncServerURL(),
			Valid:  true,
		}
	}

	return model
}

