package userpreferences

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// UserPreferencesService implements IUserPreferencesService
type UserPreferencesService struct {
	prefsRepo secondary.IUserPreferencesRepository
}

// NewUserPreferencesService creates a new UserPreferencesService instance
func NewUserPreferencesService(prefsRepo secondary.IUserPreferencesRepository) primary.IUserPreferencesService {
	return &UserPreferencesService{
		prefsRepo: prefsRepo,
	}
}

// FindByUserID finds preferences for a user
func (s *UserPreferencesService) FindByUserID(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error) {
	prefs, err := s.prefsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	if prefs == nil {
		// Create default preferences if not found
		return s.ResetToDefaults(ctx, userID)
	}
	
	return prefs, nil
}

// Update updates user preferences
func (s *UserPreferencesService) Update(ctx context.Context, userID int64, prefs *userpreferences.UserPreferences) error {
	prefs.SetUpdatedAt(time.Now())
	return s.prefsRepo.Update(ctx, userID, prefs.GetID(), prefs)
}

// ResetToDefaults resets user preferences to default values
func (s *UserPreferencesService) ResetToDefaults(ctx context.Context, userID int64) (*userpreferences.UserPreferences, error) {
	// Check if preferences already exist to get the ID
	existing, err := s.prefsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	// Default time for next day starts at 4 AM
	nextDayStartsAt := time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)

	builder := userpreferences.NewBuilder().
		WithUserID(userID).
		WithLanguage("pt-BR").
		WithTheme(valueobjects.ThemeTypeAuto).
		WithAutoSync(true).
		WithNextDayStartsAt(nextDayStartsAt).
		WithLearnAheadLimit(20).
		WithTimeboxTimeLimit(0).
		WithVideoDriver("auto").
		WithUISize(1.0).
		WithMinimalistMode(false).
		WithReduceMotion(false).
		WithPasteStripsFormatting(false).
		WithPasteImagesAsPNG(false).
		WithDefaultDeckBehavior("current_deck").
		WithShowPlayButtons(true).
		WithInterruptAudioOnAnswer(true).
		WithShowRemainingCount(true).
		WithShowNextReviewTime(false).
		WithSpacebarAnswersCard(true).
		WithIgnoreAccentsInSearch(false).
		WithSyncAudioAndImages(true).
		WithPeriodicallySyncMedia(false).
		WithForceOneWaySync(false).
		WithCreatedAt(now).
		WithUpdatedAt(now)

	if existing != nil {
		builder = builder.WithID(existing.GetID())
	}

	prefs, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build default preferences: %w", err)
	}

	if existing != nil {
		if err := s.prefsRepo.Update(ctx, userID, existing.GetID(), prefs); err != nil {
			return nil, fmt.Errorf("failed to update default preferences: %w", err)
		}
	} else {
		if err := s.prefsRepo.Save(ctx, userID, prefs); err != nil {
			return nil, fmt.Errorf("failed to save default preferences: %w", err)
		}
	}

	return prefs, nil
}

