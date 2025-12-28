package entities
import (
	"github.com/felipesantos/anki-backend/core/domain/entities"
)

import (
	"errors"
	"testing"
	"time"
)

func TestProfile_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		profile  *entities.Profile
		expected bool
	}{
		{
			name: "active profile",
			profile: &entities.Profile{
				DeletedAt: nil,
			},
			expected: true,
		},
		{
			name: "deleted profile",
			profile: &entities.Profile{
				DeletedAt: timePtr(time.Now()),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.profile.IsActive()
			if got != tt.expected {
				t.Errorf("Profile.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProfile_EnableAnkiWebSync(t *testing.T) {
	profile := &entities.Profile{
		AnkiWebSyncEnabled: false,
		AnkiWebUsername:    nil,
		UpdatedAt:         time.Now(),
	}

	// Enable with valid username
	err := profile.EnableAnkiWebSync("testuser")
	if err != nil {
		t.Errorf("Profile.EnableAnkiWebSync() error = %v, want nil", err)
	}

	if !profile.AnkiWebSyncEnabled {
		t.Errorf("Profile.EnableAnkiWebSync() failed to enable sync")
	}

	if profile.AnkiWebUsername == nil || *profile.AnkiWebUsername != "testuser" {
		t.Errorf("Profile.EnableAnkiWebSync() failed to set username")
	}

	// Try to enable with empty username
	err = profile.EnableAnkiWebSync("")
	if err == nil {
		t.Errorf("Profile.EnableAnkiWebSync() expected error for empty username")
	}
	if !errors.Is(err, entities.ErrAnkiWebUsernameEmpty) {
		t.Errorf("Profile.EnableAnkiWebSync() error = %v, want entities.ErrAnkiWebUsernameEmpty", err)
	}
}

func TestProfile_DisableAnkiWebSync(t *testing.T) {
	username := "testuser"
	profile := &entities.Profile{
		AnkiWebSyncEnabled: true,
		AnkiWebUsername:    &username,
		UpdatedAt:          time.Now(),
	}

	profile.DisableAnkiWebSync()
	if profile.AnkiWebSyncEnabled {
		t.Errorf("Profile.DisableAnkiWebSync() failed to disable sync")
	}

	if profile.AnkiWebUsername != nil {
		t.Errorf("Profile.DisableAnkiWebSync() failed to clear username")
	}
}


