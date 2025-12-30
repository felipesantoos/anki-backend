package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeckOptionsPresetModel_Creation(t *testing.T) {
	now := time.Now()
	model := &DeckOptionsPresetModel{
		ID:          1,
		UserID:      100,
		Name:        "Default Preset",
		OptionsJSON: `{"newCardsPerDay":20}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Default Preset", model.Name)
}

