package entities
import (
	"github.com/felipesantos/anki-backend/core/domain/entities"
)

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestUserPreferences_GetTheme(t *testing.T) {
	prefs := &entities.UserPreferences{
		Theme: valueobjects.ThemeTypeDark,
	}

	if prefs.GetTheme() != valueobjects.ThemeTypeDark {
		t.Errorf("UserPreferences.GetTheme() = %v, want ThemeTypeDark", prefs.GetTheme())
	}
}

func TestUserPreferences_SetTheme(t *testing.T) {
	prefs := &entities.UserPreferences{
		Theme:     valueobjects.ThemeTypeLight,
		UpdatedAt: time.Now(),
	}

	// Set valid theme
	prefs.SetTheme(valueobjects.ThemeTypeDark)
	if prefs.Theme != valueobjects.ThemeTypeDark {
		t.Errorf("UserPreferences.SetTheme() failed to set theme")
	}

	// Verify UpdatedAt was changed
	originalUpdatedAt := prefs.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	prefs.SetTheme(valueobjects.ThemeTypeAuto)
	if prefs.UpdatedAt.Equal(originalUpdatedAt) {
		t.Errorf("UserPreferences.SetTheme() should update UpdatedAt")
	}

	// Try to set invalid theme (should not change)
	prefs.SetTheme(valueobjects.ThemeType("invalid"))
	if prefs.Theme != valueobjects.ThemeTypeAuto {
		t.Errorf("UserPreferences.SetTheme() should not accept invalid theme")
	}
}

