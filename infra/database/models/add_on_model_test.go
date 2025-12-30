package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddOnModel_Creation(t *testing.T) {
	now := time.Now()
	model := &AddOnModel{
		ID:          1,
		UserID:      100,
		Code:        "1234567890",
		Name:        "Test Add-on",
		Version:     "1.0.0",
		Enabled:     true,
		InstalledAt: now,
		UpdatedAt:   now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Test Add-on", model.Name)
}

