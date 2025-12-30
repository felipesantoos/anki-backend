package mappers

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestToUserPreferencesResponse(t *testing.T) {
	now := time.Now()
	searchText := "test"
	syncURL := "http://localhost:8080"
	
	up, _ := userpreferences.NewBuilder().
		WithID(10).
		WithUserID(1).
		WithLanguage("en").
		WithTheme(valueobjects.ThemeTypeDark).
		WithAutoSync(true).
		WithNextDayStartsAt(now).
		WithLearnAheadLimit(20).
		WithTimeboxTimeLimit(30).
		WithVideoDriver("opengl").
		WithUISize(1.0).
		WithMinimalistMode(false).
		WithReduceMotion(false).
		WithPasteStripsFormatting(true).
		WithPasteImagesAsPNG(true).
		WithDefaultDeckBehavior("last").
		WithShowPlayButtons(true).
		WithInterruptAudioOnAnswer(false).
		WithShowRemainingCount(true).
		WithShowNextReviewTime(true).
		WithSpacebarAnswersCard(true).
		WithIgnoreAccentsInSearch(true).
		WithDefaultSearchText(&searchText).
		WithSyncAudioAndImages(true).
		WithPeriodicallySyncMedia(true).
		WithForceOneWaySync(false).
		WithSelfHostedSyncServerURL(&syncURL).
		Build()

	t.Run("Success", func(t *testing.T) {
		res := ToUserPreferencesResponse(up)
		assert.NotNil(t, res)
		assert.Equal(t, up.GetID(), res.ID)
		assert.Equal(t, up.GetUserID(), res.UserID)
		assert.Equal(t, up.GetLanguage(), res.Language)
		assert.Equal(t, string(up.GetTheme()), res.Theme)
		assert.Equal(t, up.GetAutoSync(), res.AutoSync)
		assert.Equal(t, up.GetNextDayStartsAt(), res.NextDayStartsAt)
		assert.Equal(t, up.GetLearnAheadLimit(), res.LearnAheadLimit)
		assert.Equal(t, up.GetTimeboxTimeLimit(), res.TimeboxTimeLimit)
		assert.Equal(t, up.GetVideoDriver(), res.VideoDriver)
		assert.Equal(t, up.GetUISize(), res.UISize)
		assert.Equal(t, up.GetMinimalistMode(), res.MinimalistMode)
		assert.Equal(t, up.GetReduceMotion(), res.ReduceMotion)
		assert.Equal(t, up.GetPasteStripsFormatting(), res.PasteStripsFormatting)
		assert.Equal(t, up.GetPasteImagesAsPNG(), res.PasteImagesAsPNG)
		assert.Equal(t, up.GetDefaultDeckBehavior(), res.DefaultDeckBehavior)
		assert.Equal(t, up.GetShowPlayButtons(), res.ShowPlayButtons)
		assert.Equal(t, up.GetInterruptAudioOnAnswer(), res.InterruptAudioOnAnswer)
		assert.Equal(t, up.GetShowRemainingCount(), res.ShowRemainingCount)
		assert.Equal(t, up.GetShowNextReviewTime(), res.ShowNextReviewTime)
		assert.Equal(t, up.GetSpacebarAnswersCard(), res.SpacebarAnswersCard)
		assert.Equal(t, up.GetIgnoreAccentsInSearch(), res.IgnoreAccentsInSearch)
		assert.Equal(t, up.GetDefaultSearchText(), res.DefaultSearchText)
		assert.Equal(t, up.GetSyncAudioAndImages(), res.SyncAudioAndImages)
		assert.Equal(t, up.GetPeriodicallySyncMedia(), res.PeriodicallySyncMedia)
		assert.Equal(t, up.GetForceOneWaySync(), res.ForceOneWaySync)
		assert.Equal(t, up.GetSelfHostedSyncServerURL(), res.SelfHostedSyncServerURL)
	})

	t.Run("NilEntity", func(t *testing.T) {
		res := ToUserPreferencesResponse(nil)
		assert.Nil(t, res)
	})
}
