package entities

import (
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

// UserPreferences represents user preferences entity in the domain
// It stores global user settings and preferences
type UserPreferences struct {
	ID                        int64
	UserID                    int64 // Unique
	Language                  string
	Theme                     valueobjects.ThemeType
	AutoSync                  bool
	NextDayStartsAt           time.Time // Time of day
	LearnAheadLimit           int       // Minutes
	TimeboxTimeLimit          int       // Minutes (0 = disabled)
	VideoDriver               string
	UISize                    float64
	MinimalistMode            bool
	ReduceMotion              bool
	PasteStripsFormatting     bool
	PasteImagesAsPNG          bool
	DefaultDeckBehavior       string
	ShowPlayButtons           bool
	InterruptAudioOnAnswer    bool
	ShowRemainingCount        bool
	ShowNextReviewTime        bool
	SpacebarAnswersCard       bool
	IgnoreAccentsInSearch     bool
	DefaultSearchText         *string
	SyncAudioAndImages        bool
	PeriodicallySyncMedia     bool
	ForceOneWaySync           bool
	SelfHostedSyncServerURL   *string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// GetTheme returns the theme value object
func (up *UserPreferences) GetTheme() valueobjects.ThemeType {
	return up.Theme
}

// SetTheme sets the theme value object
func (up *UserPreferences) SetTheme(theme valueobjects.ThemeType) {
	if theme.IsValid() {
		up.Theme = theme
		up.UpdatedAt = time.Now()
	}
}

