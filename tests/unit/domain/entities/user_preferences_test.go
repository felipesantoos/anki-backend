package entities
import (
	userpreferences "github.com/felipesantos/anki-backend/core/domain/entities/user_preferences"
)

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestUserPreferences_GetTheme(t *testing.T) {
	prefs := &userpreferences.UserPreferences{}
	prefs.SetTheme(valueobjects.ThemeTypeDark)

	if prefs.GetTheme() != valueobjects.ThemeTypeDark {
		t.Errorf("UserPreferences.GetTheme() = %v, want ThemeTypeDark", prefs.GetTheme())
	}
}

func TestUserPreferences_SetTheme(t *testing.T) {
	prefs := &userpreferences.UserPreferences{}
	prefs.SetTheme(valueobjects.ThemeTypeLight)
	prefs.SetUpdatedAt(time.Now())

	// Set valid theme
	prefs.SetTheme(valueobjects.ThemeTypeDark)
	if prefs.GetTheme() != valueobjects.ThemeTypeDark {
		t.Errorf("UserPreferences.SetTheme() failed to set theme")
	}

	// Verify UpdatedAt was changed
	originalUpdatedAt := prefs.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)
	prefs.SetTheme(valueobjects.ThemeTypeAuto)
	if prefs.GetUpdatedAt().Equal(originalUpdatedAt) {
		t.Errorf("UserPreferences.SetTheme() should update UpdatedAt")
	}

	// Try to set invalid theme (should not change)
	prefs.SetTheme(valueobjects.ThemeType("invalid"))
	if prefs.GetTheme() != valueobjects.ThemeTypeAuto {
		t.Errorf("UserPreferences.SetTheme() should not accept invalid theme")
	}
}

