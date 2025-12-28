package entities

import (
	"testing"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities"
)

func TestAddOn_IsEnabled(t *testing.T) {
	enabledAddOn := &entities.AddOn{}
	enabledAddOn.SetEnabled(true)

	if !enabledAddOn.IsEnabled() {
		t.Errorf("AddOn.IsEnabled() = false, want true")
	}

	disabledAddOn := &entities.AddOn{}
	disabledAddOn.SetEnabled(false)

	if disabledAddOn.IsEnabled() {
		t.Errorf("AddOn.IsEnabled() = true, want false")
	}
}

func TestAddOn_Enable(t *testing.T) {
	addOn := &entities.AddOn{}
	addOn.SetEnabled(false)
	addOn.SetUpdatedAt(time.Now())

	addOn.Enable()
	if !addOn.GetEnabled() {
		t.Errorf("AddOn.Enable() failed to enable add-on")
	}

	// Enable again (should be idempotent)
	addOn.Enable()
	if !addOn.GetEnabled() {
		t.Errorf("AddOn.Enable() failed to keep add-on enabled")
	}
}

func TestAddOn_Disable(t *testing.T) {
	addOn := &entities.AddOn{}
	addOn.SetEnabled(true)
	addOn.SetUpdatedAt(time.Now())

	addOn.Disable()
	if addOn.GetEnabled() {
		t.Errorf("AddOn.Disable() failed to disable add-on")
	}

	// Disable again (should be idempotent)
	addOn.Disable()
	if addOn.GetEnabled() {
		t.Errorf("AddOn.Disable() failed to keep add-on disabled")
	}
}

