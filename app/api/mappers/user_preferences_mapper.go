package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
)

// ToUserPreferencesResponse converts UserPreferences domain entity to Response DTO
func ToUserPreferencesResponse(up *userpreferences.UserPreferences) *response.UserPreferencesResponse {
	if up == nil {
		return nil
	}
	return &response.UserPreferencesResponse{
		ID:                      up.GetID(),
		UserID:                  up.GetUserID(),
		Language:                up.GetLanguage(),
		Theme:                   string(up.GetTheme()),
		AutoSync:                up.GetAutoSync(),
		NextDayStartsAt:         up.GetNextDayStartsAt(),
		LearnAheadLimit:         up.GetLearnAheadLimit(),
		TimeboxTimeLimit:        up.GetTimeboxTimeLimit(),
		VideoDriver:             up.GetVideoDriver(),
		UISize:                  up.GetUISize(),
		MinimalistMode:          up.GetMinimalistMode(),
		ReduceMotion:            up.GetReduceMotion(),
		PasteStripsFormatting:   up.GetPasteStripsFormatting(),
		PasteImagesAsPNG:        up.GetPasteImagesAsPNG(),
		DefaultDeckBehavior:     up.GetDefaultDeckBehavior(),
		ShowPlayButtons:         up.GetShowPlayButtons(),
		InterruptAudioOnAnswer:  up.GetInterruptAudioOnAnswer(),
		ShowRemainingCount:      up.GetShowRemainingCount(),
		ShowNextReviewTime:      up.GetShowNextReviewTime(),
		SpacebarAnswersCard:     up.GetSpacebarAnswersCard(),
		IgnoreAccentsInSearch:   up.GetIgnoreAccentsInSearch(),
		DefaultSearchText:       up.GetDefaultSearchText(),
		SyncAudioAndImages:      up.GetSyncAudioAndImages(),
		PeriodicallySyncMedia:   up.GetPeriodicallySyncMedia(),
		ForceOneWaySync:         up.GetForceOneWaySync(),
		SelfHostedSyncServerURL: up.GetSelfHostedSyncServerURL(),
	}
}

