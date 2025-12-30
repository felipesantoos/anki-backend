package response

import "time"

// UserPreferencesResponse represents the response payload for user preferences
type UserPreferencesResponse struct {
	ID                      int64     `json:"id"`
	UserID                  int64     `json:"user_id"`
	Language                string    `json:"language"`
	Theme                   string    `json:"theme"`
	AutoSync                bool      `json:"auto_sync"`
	NextDayStartsAt         time.Time `json:"next_day_starts_at"`
	LearnAheadLimit         int       `json:"learn_ahead_limit"`
	TimeboxTimeLimit        int       `json:"timebox_time_limit"`
	VideoDriver             string    `json:"video_driver"`
	UISize                  float64   `json:"ui_size"`
	MinimalistMode          bool      `json:"minimalist_mode"`
	ReduceMotion            bool      `json:"reduce_motion"`
	PasteStripsFormatting   bool      `json:"paste_strips_formatting"`
	PasteImagesAsPNG        bool      `json:"paste_images_as_png"`
	DefaultDeckBehavior     string    `json:"default_deck_behavior"`
	ShowPlayButtons         bool      `json:"show_play_buttons"`
	InterruptAudioOnAnswer  bool      `json:"interrupt_audio_on_answer"`
	ShowRemainingCount      bool      `json:"show_remaining_count"`
	ShowNextReviewTime      bool      `json:"show_next_review_time"`
	SpacebarAnswersCard     bool      `json:"spacebar_answers_card"`
	IgnoreAccentsInSearch   bool      `json:"ignore_accents_in_search"`
	DefaultSearchText       *string   `json:"default_search_text"`
	SyncAudioAndImages      bool      `json:"sync_audio_and_images"`
	PeriodicallySyncMedia   bool      `json:"periodically_sync_media"`
	ForceOneWaySync         bool      `json:"force_one_way_sync"`
	SelfHostedSyncServerURL *string   `json:"self_hosted_sync_server_url"`
}

