package entities
import (
	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
)

import (
	"errors"
	"testing"
	"time"
)

func TestProfile_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		profile  *profile.Profile
		expected bool
	}{
		{
			name: "active profile",
			profile: func() *profile.Profile {
				p := &profile.Profile{}
				p.SetDeletedAt(nil)
				return p
			}(),
			expected: true,
		},
		{
			name: "deleted profile",
			profile: func() *profile.Profile {
				p := &profile.Profile{}
				p.SetDeletedAt(timePtr(time.Now()))
				return p
			}(),
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
	p := &profile.Profile{}
	p.SetAnkiWebSyncEnabled(false)
	p.SetAnkiWebUsername(nil)
	p.SetUpdatedAt(time.Now())

	// Enable with valid username
	err := p.EnableAnkiWebSync("testuser")
	if err != nil {
		t.Errorf("Profile.EnableAnkiWebSync() error = %v, want nil", err)
	}

	if !p.GetAnkiWebSyncEnabled() {
		t.Errorf("Profile.EnableAnkiWebSync() failed to enable sync")
	}

	if p.GetAnkiWebUsername() == nil || *p.GetAnkiWebUsername() != "testuser" {
		t.Errorf("Profile.EnableAnkiWebSync() failed to set username")
	}

	// Try to enable with empty username
	err = p.EnableAnkiWebSync("")
	if err == nil {
		t.Errorf("Profile.EnableAnkiWebSync() expected error for empty username")
	}
	if !errors.Is(err, profile.ErrAnkiWebUsernameEmpty) {
		t.Errorf("Profile.EnableAnkiWebSync() error = %v, want profile.ErrAnkiWebUsernameEmpty", err)
	}
}

func TestProfile_DisableAnkiWebSync(t *testing.T) {
	username := "testuser"
	p := &profile.Profile{}
	p.SetAnkiWebSyncEnabled(true)
	p.SetAnkiWebUsername(&username)
	p.SetUpdatedAt(time.Now())

	p.DisableAnkiWebSync()
	if p.GetAnkiWebSyncEnabled() {
		t.Errorf("Profile.DisableAnkiWebSync() failed to disable sync")
	}

	if p.GetAnkiWebUsername() != nil {
		t.Errorf("Profile.DisableAnkiWebSync() failed to clear username")
	}
}


