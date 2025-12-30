package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// UserPreferencesRepository implements IUserPreferencesRepository using PostgreSQL
type UserPreferencesRepository struct {
	db *sql.DB
}

// NewUserPreferencesRepository creates a new UserPreferencesRepository instance
func NewUserPreferencesRepository(db *sql.DB) secondary.IUserPreferencesRepository {
	return &UserPreferencesRepository{
		db: db,
	}
}

// Save saves or updates user preferences in the database
func (r *UserPreferencesRepository) Save(ctx context.Context, userID int64, prefsEntity *userpreferences.UserPreferences) error {
	model := mappers.UserPreferencesToModel(prefsEntity)

	if prefsEntity.GetID() == 0 {
		// Insert new preferences
		query := `
			INSERT INTO user_preferences (
				user_id, language, theme, auto_sync, next_day_starts_at, learn_ahead_limit,
				timebox_time_limit, video_driver, ui_size, minimalist_mode, reduce_motion,
				paste_strips_formatting, paste_images_as_png, default_deck_behavior,
				show_play_buttons, interrupt_audio_on_answer, show_remaining_count,
				show_next_review_time, spacebar_answers_card, ignore_accents_in_search,
				default_search_text, sync_audio_and_images, periodically_sync_media,
				force_one_way_sync, self_hosted_sync_server_url, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var defaultSearchText interface{}
		if model.DefaultSearchText.Valid {
			defaultSearchText = model.DefaultSearchText.String
		}

		var selfHostedURL interface{}
		if model.SelfHostedSyncServerURL.Valid {
			selfHostedURL = model.SelfHostedSyncServerURL.String
		}

		var prefsID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Language,
			model.Theme,
			model.AutoSync,
			model.NextDayStartsAt.Format("15:04:05"), // Format as TIME string
			model.LearnAheadLimit,
			model.TimeboxTimeLimit,
			model.VideoDriver,
			model.UISize,
			model.MinimalistMode,
			model.ReduceMotion,
			model.PasteStripsFormatting,
			model.PasteImagesAsPNG,
			model.DefaultDeckBehavior,
			model.ShowPlayButtons,
			model.InterruptAudioOnAnswer,
			model.ShowRemainingCount,
			model.ShowNextReviewTime,
			model.SpacebarAnswersCard,
			model.IgnoreAccentsInSearch,
			defaultSearchText,
			model.SyncAudioAndImages,
			model.PeriodicallySyncMedia,
			model.ForceOneWaySync,
			selfHostedURL,
			model.CreatedAt,
			model.UpdatedAt,
		).Scan(&prefsID)
		if err != nil {
			return fmt.Errorf("failed to create user preferences: %w", err)
		}

		prefsEntity.SetID(prefsID)
		return nil
	}

	// Update existing preferences - validate ownership first
	existingPrefs, err := r.FindByID(ctx, userID, prefsEntity.GetID())
	if err != nil {
		return err
	}
	if existingPrefs == nil {
		return ownership.ErrResourceNotFound
	}

	// Update preferences
	query := `
		UPDATE user_preferences
		SET language = $1, theme = $2, auto_sync = $3, next_day_starts_at = $4,
			learn_ahead_limit = $5, timebox_time_limit = $6, video_driver = $7,
			ui_size = $8, minimalist_mode = $9, reduce_motion = $10,
			paste_strips_formatting = $11, paste_images_as_png = $12,
			default_deck_behavior = $13, show_play_buttons = $14,
			interrupt_audio_on_answer = $15, show_remaining_count = $16,
			show_next_review_time = $17, spacebar_answers_card = $18,
			ignore_accents_in_search = $19, default_search_text = $20,
			sync_audio_and_images = $21, periodically_sync_media = $22,
			force_one_way_sync = $23, self_hosted_sync_server_url = $24,
			updated_at = $25
		WHERE id = $26 AND user_id = $27
	`

	now := time.Now()
	model.UpdatedAt = now

	var defaultSearchText interface{}
	if model.DefaultSearchText.Valid {
		defaultSearchText = model.DefaultSearchText.String
	}

	var selfHostedURL interface{}
	if model.SelfHostedSyncServerURL.Valid {
		selfHostedURL = model.SelfHostedSyncServerURL.String
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Language,
		model.Theme,
		model.AutoSync,
		model.NextDayStartsAt.Format("15:04:05"), // Format as TIME
		model.LearnAheadLimit,
		model.TimeboxTimeLimit,
		model.VideoDriver,
		model.UISize,
		model.MinimalistMode,
		model.ReduceMotion,
		model.PasteStripsFormatting,
		model.PasteImagesAsPNG,
		model.DefaultDeckBehavior,
		model.ShowPlayButtons,
		model.InterruptAudioOnAnswer,
		model.ShowRemainingCount,
		model.ShowNextReviewTime,
		model.SpacebarAnswersCard,
		model.IgnoreAccentsInSearch,
		defaultSearchText,
		model.SyncAudioAndImages,
		model.PeriodicallySyncMedia,
		model.ForceOneWaySync,
		selfHostedURL,
		model.UpdatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// FindByID finds user preferences by ID, filtering by userID to ensure ownership
func (r *UserPreferencesRepository) FindByID(ctx context.Context, userID int64, id int64) (*userpreferences.UserPreferences, error) {
	query := `
		SELECT id, user_id, language, theme, auto_sync, next_day_starts_at,
			learn_ahead_limit, timebox_time_limit, video_driver, ui_size,
			minimalist_mode, reduce_motion, paste_strips_formatting, paste_images_as_png,
			default_deck_behavior, show_play_buttons, interrupt_audio_on_answer,
			show_remaining_count, show_next_review_time, spacebar_answers_card,
			ignore_accents_in_search, default_search_text, sync_audio_and_images,
			periodically_sync_media, force_one_way_sync, self_hosted_sync_server_url,
			created_at, updated_at
		FROM user_preferences
		WHERE id = $1 AND user_id = $2
	`

	var model models.UserPreferencesModel
	var defaultSearchText, selfHostedURL sql.NullString
	var nextDayTime time.Time

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Language,
		&model.Theme,
		&model.AutoSync,
		&nextDayTime,
		&model.LearnAheadLimit,
		&model.TimeboxTimeLimit,
		&model.VideoDriver,
		&model.UISize,
		&model.MinimalistMode,
		&model.ReduceMotion,
		&model.PasteStripsFormatting,
		&model.PasteImagesAsPNG,
		&model.DefaultDeckBehavior,
		&model.ShowPlayButtons,
		&model.InterruptAudioOnAnswer,
		&model.ShowRemainingCount,
		&model.ShowNextReviewTime,
		&model.SpacebarAnswersCard,
		&model.IgnoreAccentsInSearch,
		&defaultSearchText,
		&model.SyncAudioAndImages,
		&model.PeriodicallySyncMedia,
		&model.ForceOneWaySync,
		&selfHostedURL,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find user preferences: %w", err)
	}

	model.NextDayStartsAt = time.Date(1970, 1, 1, nextDayTime.Hour(), nextDayTime.Minute(), nextDayTime.Second(), 0, time.UTC)

	model.DefaultSearchText = defaultSearchText
	model.SelfHostedSyncServerURL = selfHostedURL

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.UserPreferencesToDomain(&model)
}

// FindByUserID finds user preferences for a user (one-to-one relationship)
func (r *UserPreferencesRepository) FindByUserID(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error) {
	query := `
		SELECT id, user_id, language, theme, auto_sync, next_day_starts_at,
			learn_ahead_limit, timebox_time_limit, video_driver, ui_size,
			minimalist_mode, reduce_motion, paste_strips_formatting, paste_images_as_png,
			default_deck_behavior, show_play_buttons, interrupt_audio_on_answer,
			show_remaining_count, show_next_review_time, spacebar_answers_card,
			ignore_accents_in_search, default_search_text, sync_audio_and_images,
			periodically_sync_media, force_one_way_sync, self_hosted_sync_server_url,
			created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1
	`

	var model models.UserPreferencesModel
	var defaultSearchText, selfHostedURL sql.NullString
	var nextDayTime time.Time

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Language,
		&model.Theme,
		&model.AutoSync,
		&nextDayTime,
		&model.LearnAheadLimit,
		&model.TimeboxTimeLimit,
		&model.VideoDriver,
		&model.UISize,
		&model.MinimalistMode,
		&model.ReduceMotion,
		&model.PasteStripsFormatting,
		&model.PasteImagesAsPNG,
		&model.DefaultDeckBehavior,
		&model.ShowPlayButtons,
		&model.InterruptAudioOnAnswer,
		&model.ShowRemainingCount,
		&model.ShowNextReviewTime,
		&model.SpacebarAnswersCard,
		&model.IgnoreAccentsInSearch,
		&defaultSearchText,
		&model.SyncAudioAndImages,
		&model.PeriodicallySyncMedia,
		&model.ForceOneWaySync,
		&selfHostedURL,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user preferences by user ID: %w", err)
	}

	model.NextDayStartsAt = time.Date(1970, 1, 1, nextDayTime.Hour(), nextDayTime.Minute(), nextDayTime.Second(), 0, time.UTC)

	model.DefaultSearchText = defaultSearchText
	model.SelfHostedSyncServerURL = selfHostedURL

	return mappers.UserPreferencesToDomain(&model)
}

// Update updates existing user preferences, validating ownership
func (r *UserPreferencesRepository) Update(ctx context.Context, userID int64, id int64, prefsEntity *userpreferences.UserPreferences) error {
	return r.Save(ctx, userID, prefsEntity)
}

// Delete deletes user preferences, validating ownership
func (r *UserPreferencesRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingPrefs, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingPrefs == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (user_preferences doesn't have soft delete)
	query := `DELETE FROM user_preferences WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user preferences: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// Exists checks if user preferences exist for a user
func (r *UserPreferencesRepository) Exists(ctx context.Context, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_preferences
			WHERE user_id = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user preferences existence: %w", err)
	}

	return exists, nil
}

// Ensure UserPreferencesRepository implements IUserPreferencesRepository
var _ secondary.IUserPreferencesRepository = (*UserPreferencesRepository)(nil)

