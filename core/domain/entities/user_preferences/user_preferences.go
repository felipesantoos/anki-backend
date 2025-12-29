package userpreferences

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// UserPreferences represents user preferences entity in the domain
// It stores global user settings and preferences
type UserPreferences struct {
	id                        int64
	userID                    int64 // Unique
	language                  string
	theme                     valueobjects.ThemeType
	autoSync                  bool
	nextDayStartsAt           time.Time // Time of day
	learnAheadLimit           int       // Minutes
	timeboxTimeLimit          int       // Minutes (0 = disabled)
	videoDriver               string
	uiSize                    float64
	minimalistMode            bool
	reduceMotion              bool
	pasteStripsFormatting     bool
	pasteImagesAsPNG          bool
	defaultDeckBehavior       string
	showPlayButtons           bool
	interruptAudioOnAnswer    bool
	showRemainingCount        bool
	showNextReviewTime        bool
	spacebarAnswersCard       bool
	ignoreAccentsInSearch     bool
	defaultSearchText         *string
	syncAudioAndImages        bool
	periodicallySyncMedia     bool
	forceOneWaySync           bool
	selfHostedSyncServerURL   *string
	createdAt                 time.Time
	updatedAt                 time.Time
}

// Getters
func (up *UserPreferences) GetID() int64 {
	return up.id
}

func (up *UserPreferences) GetUserID() int64 {
	return up.userID
}

func (up *UserPreferences) GetLanguage() string {
	return up.language
}

func (up *UserPreferences) GetAutoSync() bool {
	return up.autoSync
}

func (up *UserPreferences) GetNextDayStartsAt() time.Time {
	return up.nextDayStartsAt
}

func (up *UserPreferences) GetLearnAheadLimit() int {
	return up.learnAheadLimit
}

func (up *UserPreferences) GetTimeboxTimeLimit() int {
	return up.timeboxTimeLimit
}

func (up *UserPreferences) GetVideoDriver() string {
	return up.videoDriver
}

func (up *UserPreferences) GetUISize() float64 {
	return up.uiSize
}

func (up *UserPreferences) GetMinimalistMode() bool {
	return up.minimalistMode
}

func (up *UserPreferences) GetReduceMotion() bool {
	return up.reduceMotion
}

func (up *UserPreferences) GetPasteStripsFormatting() bool {
	return up.pasteStripsFormatting
}

func (up *UserPreferences) GetPasteImagesAsPNG() bool {
	return up.pasteImagesAsPNG
}

func (up *UserPreferences) GetDefaultDeckBehavior() string {
	return up.defaultDeckBehavior
}

func (up *UserPreferences) GetShowPlayButtons() bool {
	return up.showPlayButtons
}

func (up *UserPreferences) GetInterruptAudioOnAnswer() bool {
	return up.interruptAudioOnAnswer
}

func (up *UserPreferences) GetShowRemainingCount() bool {
	return up.showRemainingCount
}

func (up *UserPreferences) GetShowNextReviewTime() bool {
	return up.showNextReviewTime
}

func (up *UserPreferences) GetSpacebarAnswersCard() bool {
	return up.spacebarAnswersCard
}

func (up *UserPreferences) GetIgnoreAccentsInSearch() bool {
	return up.ignoreAccentsInSearch
}

func (up *UserPreferences) GetDefaultSearchText() *string {
	return up.defaultSearchText
}

func (up *UserPreferences) GetSyncAudioAndImages() bool {
	return up.syncAudioAndImages
}

func (up *UserPreferences) GetPeriodicallySyncMedia() bool {
	return up.periodicallySyncMedia
}

func (up *UserPreferences) GetForceOneWaySync() bool {
	return up.forceOneWaySync
}

func (up *UserPreferences) GetSelfHostedSyncServerURL() *string {
	return up.selfHostedSyncServerURL
}

func (up *UserPreferences) GetCreatedAt() time.Time {
	return up.createdAt
}

func (up *UserPreferences) GetUpdatedAt() time.Time {
	return up.updatedAt
}

// Setters
func (up *UserPreferences) SetID(id int64) {
	up.id = id
}

func (up *UserPreferences) SetUserID(userID int64) {
	up.userID = userID
}

func (up *UserPreferences) SetLanguage(language string) {
	up.language = language
}

func (up *UserPreferences) SetAutoSync(autoSync bool) {
	up.autoSync = autoSync
}

func (up *UserPreferences) SetNextDayStartsAt(nextDayStartsAt time.Time) {
	up.nextDayStartsAt = nextDayStartsAt
}

func (up *UserPreferences) SetLearnAheadLimit(learnAheadLimit int) {
	up.learnAheadLimit = learnAheadLimit
}

func (up *UserPreferences) SetTimeboxTimeLimit(timeboxTimeLimit int) {
	up.timeboxTimeLimit = timeboxTimeLimit
}

func (up *UserPreferences) SetVideoDriver(videoDriver string) {
	up.videoDriver = videoDriver
}

func (up *UserPreferences) SetUISize(uiSize float64) {
	up.uiSize = uiSize
}

func (up *UserPreferences) SetMinimalistMode(minimalistMode bool) {
	up.minimalistMode = minimalistMode
}

func (up *UserPreferences) SetReduceMotion(reduceMotion bool) {
	up.reduceMotion = reduceMotion
}

func (up *UserPreferences) SetPasteStripsFormatting(pasteStripsFormatting bool) {
	up.pasteStripsFormatting = pasteStripsFormatting
}

func (up *UserPreferences) SetPasteImagesAsPNG(pasteImagesAsPNG bool) {
	up.pasteImagesAsPNG = pasteImagesAsPNG
}

func (up *UserPreferences) SetDefaultDeckBehavior(defaultDeckBehavior string) {
	up.defaultDeckBehavior = defaultDeckBehavior
}

func (up *UserPreferences) SetShowPlayButtons(showPlayButtons bool) {
	up.showPlayButtons = showPlayButtons
}

func (up *UserPreferences) SetInterruptAudioOnAnswer(interruptAudioOnAnswer bool) {
	up.interruptAudioOnAnswer = interruptAudioOnAnswer
}

func (up *UserPreferences) SetShowRemainingCount(showRemainingCount bool) {
	up.showRemainingCount = showRemainingCount
}

func (up *UserPreferences) SetShowNextReviewTime(showNextReviewTime bool) {
	up.showNextReviewTime = showNextReviewTime
}

func (up *UserPreferences) SetSpacebarAnswersCard(spacebarAnswersCard bool) {
	up.spacebarAnswersCard = spacebarAnswersCard
}

func (up *UserPreferences) SetIgnoreAccentsInSearch(ignoreAccentsInSearch bool) {
	up.ignoreAccentsInSearch = ignoreAccentsInSearch
}

func (up *UserPreferences) SetDefaultSearchText(defaultSearchText *string) {
	up.defaultSearchText = defaultSearchText
}

func (up *UserPreferences) SetSyncAudioAndImages(syncAudioAndImages bool) {
	up.syncAudioAndImages = syncAudioAndImages
}

func (up *UserPreferences) SetPeriodicallySyncMedia(periodicallySyncMedia bool) {
	up.periodicallySyncMedia = periodicallySyncMedia
}

func (up *UserPreferences) SetForceOneWaySync(forceOneWaySync bool) {
	up.forceOneWaySync = forceOneWaySync
}

func (up *UserPreferences) SetSelfHostedSyncServerURL(selfHostedSyncServerURL *string) {
	up.selfHostedSyncServerURL = selfHostedSyncServerURL
}

func (up *UserPreferences) SetCreatedAt(createdAt time.Time) {
	up.createdAt = createdAt
}

func (up *UserPreferences) SetUpdatedAt(updatedAt time.Time) {
	up.updatedAt = updatedAt
}

// GetTheme returns the theme value object
func (up *UserPreferences) GetTheme() valueobjects.ThemeType {
	return up.theme
}

// SetTheme sets the theme value object
func (up *UserPreferences) SetTheme(theme valueobjects.ThemeType) {
	if theme.IsValid() {
		up.theme = theme
		up.updatedAt = time.Now()
	}
}

