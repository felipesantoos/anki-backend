package userpreferences

import (
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

var (
	ErrUserIDRequired = errors.New("userID is required")
	ErrInvalidTheme   = errors.New("invalid theme type")
)

type UserPreferencesBuilder struct {
	userPreferences *UserPreferences
	errs            []error // Lista de erros acumulados
}

func NewBuilder() *UserPreferencesBuilder {
	return &UserPreferencesBuilder{
		userPreferences: &UserPreferences{},
		errs:            make([]error, 0),
	}
}

func (b *UserPreferencesBuilder) WithID(id int64) *UserPreferencesBuilder {
	if id < 0 {
		b.errs = append(b.errs, errors.New("id must be non-negative"))
		return b
	}
	b.userPreferences.id = id // Acesso direto ao campo privado (mesmo package)
	return b
}

func (b *UserPreferencesBuilder) WithUserID(userID int64) *UserPreferencesBuilder {
	if userID <= 0 {
		b.errs = append(b.errs, ErrUserIDRequired)
		return b
	}
	b.userPreferences.userID = userID // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithLanguage(language string) *UserPreferencesBuilder {
	b.userPreferences.language = language // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithTheme(theme valueobjects.ThemeType) *UserPreferencesBuilder {
	if !theme.IsValid() {
		b.errs = append(b.errs, ErrInvalidTheme)
		return b
	}
	b.userPreferences.theme = theme // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithAutoSync(autoSync bool) *UserPreferencesBuilder {
	b.userPreferences.autoSync = autoSync // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithNextDayStartsAt(nextDayStartsAt time.Time) *UserPreferencesBuilder {
	b.userPreferences.nextDayStartsAt = nextDayStartsAt // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithLearnAheadLimit(learnAheadLimit int) *UserPreferencesBuilder {
	b.userPreferences.learnAheadLimit = learnAheadLimit // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithTimeboxTimeLimit(timeboxTimeLimit int) *UserPreferencesBuilder {
	b.userPreferences.timeboxTimeLimit = timeboxTimeLimit // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithVideoDriver(videoDriver string) *UserPreferencesBuilder {
	b.userPreferences.videoDriver = videoDriver // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithUISize(uiSize float64) *UserPreferencesBuilder {
	b.userPreferences.uiSize = uiSize // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithMinimalistMode(minimalistMode bool) *UserPreferencesBuilder {
	b.userPreferences.minimalistMode = minimalistMode // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithReduceMotion(reduceMotion bool) *UserPreferencesBuilder {
	b.userPreferences.reduceMotion = reduceMotion // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithPasteStripsFormatting(pasteStripsFormatting bool) *UserPreferencesBuilder {
	b.userPreferences.pasteStripsFormatting = pasteStripsFormatting // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithPasteImagesAsPNG(pasteImagesAsPNG bool) *UserPreferencesBuilder {
	b.userPreferences.pasteImagesAsPNG = pasteImagesAsPNG // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithDefaultDeckBehavior(defaultDeckBehavior string) *UserPreferencesBuilder {
	b.userPreferences.defaultDeckBehavior = defaultDeckBehavior // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithShowPlayButtons(showPlayButtons bool) *UserPreferencesBuilder {
	b.userPreferences.showPlayButtons = showPlayButtons // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithInterruptAudioOnAnswer(interruptAudioOnAnswer bool) *UserPreferencesBuilder {
	b.userPreferences.interruptAudioOnAnswer = interruptAudioOnAnswer // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithShowRemainingCount(showRemainingCount bool) *UserPreferencesBuilder {
	b.userPreferences.showRemainingCount = showRemainingCount // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithShowNextReviewTime(showNextReviewTime bool) *UserPreferencesBuilder {
	b.userPreferences.showNextReviewTime = showNextReviewTime // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithSpacebarAnswersCard(spacebarAnswersCard bool) *UserPreferencesBuilder {
	b.userPreferences.spacebarAnswersCard = spacebarAnswersCard // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithIgnoreAccentsInSearch(ignoreAccentsInSearch bool) *UserPreferencesBuilder {
	b.userPreferences.ignoreAccentsInSearch = ignoreAccentsInSearch // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithDefaultSearchText(defaultSearchText *string) *UserPreferencesBuilder {
	b.userPreferences.defaultSearchText = defaultSearchText // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithSyncAudioAndImages(syncAudioAndImages bool) *UserPreferencesBuilder {
	b.userPreferences.syncAudioAndImages = syncAudioAndImages // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithPeriodicallySyncMedia(periodicallySyncMedia bool) *UserPreferencesBuilder {
	b.userPreferences.periodicallySyncMedia = periodicallySyncMedia // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithForceOneWaySync(forceOneWaySync bool) *UserPreferencesBuilder {
	b.userPreferences.forceOneWaySync = forceOneWaySync // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithSelfHostedSyncServerURL(selfHostedSyncServerURL *string) *UserPreferencesBuilder {
	b.userPreferences.selfHostedSyncServerURL = selfHostedSyncServerURL // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithCreatedAt(createdAt time.Time) *UserPreferencesBuilder {
	b.userPreferences.createdAt = createdAt // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) WithUpdatedAt(updatedAt time.Time) *UserPreferencesBuilder {
	b.userPreferences.updatedAt = updatedAt // Acesso direto ao campo privado
	return b
}

func (b *UserPreferencesBuilder) Build() (*UserPreferences, error) {
	if len(b.errs) > 0 {
		// Retornar todos os erros acumulados
		return nil, fmt.Errorf("validation errors: %v", b.errs)
	}
	return b.userPreferences, nil
}

// HasErrors retorna true se há erros acumulados
func (b *UserPreferencesBuilder) HasErrors() bool {
	return len(b.errs) > 0
}

// Errors retorna a lista de erros acumulados
func (b *UserPreferencesBuilder) Errors() []error {
	return b.errs
}

