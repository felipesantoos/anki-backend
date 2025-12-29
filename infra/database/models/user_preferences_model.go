package models

import (
	"database/sql"
	"time"
)

// UserPreferencesModel represents the user_preferences table structure in the database
type UserPreferencesModel struct {
	ID                         int64
	UserID                     int64
	Language                   string
	Theme                      string // theme_type enum
	AutoSync                   bool
	NextDayStartsAt            time.Time // TIME stored as time.Time (using date part as 1970-01-01)
	LearnAheadLimit            int
	TimeboxTimeLimit           int
	VideoDriver                string
	UISize                     float64
	MinimalistMode             bool
	ReduceMotion               bool
	PasteStripsFormatting      bool
	PasteImagesAsPNG           bool
	DefaultDeckBehavior        string
	ShowPlayButtons            bool
	InterruptAudioOnAnswer     bool
	ShowRemainingCount         bool
	ShowNextReviewTime         bool
	SpacebarAnswersCard        bool
	IgnoreAccentsInSearch      bool
	DefaultSearchText          sql.NullString
	SyncAudioAndImages         bool
	PeriodicallySyncMedia      bool
	ForceOneWaySync            bool
	SelfHostedSyncServerURL    sql.NullString
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

