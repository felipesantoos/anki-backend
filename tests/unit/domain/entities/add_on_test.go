package entities

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
)

func TestAddOn_IsEnabled(t *testing.T) {
	enabledAddOn := &entities.AddOn{
		Enabled: true,
	}

	if !enabledAddOn.IsEnabled() {
		t.Errorf("AddOn.IsEnabled() = false, want true")
	}

	disabledAddOn := &entities.AddOn{
		Enabled: false,
	}

	if disabledAddOn.IsEnabled() {
		t.Errorf("AddOn.IsEnabled() = true, want false")
	}
}

func TestAddOn_Enable(t *testing.T) {
	addOn := &entities.AddOn{
		Enabled:   false,
		UpdatedAt: time.Now(),
	}

	addOn.Enable()
	if !addOn.Enabled {
		t.Errorf("AddOn.Enable() failed to enable add-on")
	}

	// Enable again (should be idempotent)
	addOn.Enable()
	if !addOn.Enabled {
		t.Errorf("AddOn.Enable() failed to keep add-on enabled")
	}
}

func TestAddOn_Disable(t *testing.T) {
	addOn := &entities.AddOn{
		Enabled:   true,
		UpdatedAt: time.Now(),
	}

	addOn.Disable()
	if addOn.Enabled {
		t.Errorf("AddOn.Disable() failed to disable add-on")
	}

	// Disable again (should be idempotent)
	addOn.Disable()
	if addOn.Enabled {
		t.Errorf("AddOn.Disable() failed to keep add-on disabled")
	}
}

