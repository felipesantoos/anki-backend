package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserPreferencesModel_Creation(t *testing.T) {
	now := time.Now()
	nextDayTime := time.Date(1970, 1, 1, 4, 0, 0, 0, time.UTC)

	model := &UserPreferencesModel{
		ID:               1,
		UserID:           100,
		Language:         "pt-BR",
		Theme:            "light",
		AutoSync:         true,
		NextDayStartsAt:  nextDayTime,
		LearnAheadLimit:  20,
		DefaultSearchText: sqlNullString("default", true),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "pt-BR", model.Language)
	assert.True(t, model.DefaultSearchText.Valid)
}

func TestUserPreferencesModel_NullFields(t *testing.T) {
	model := &UserPreferencesModel{
		ID:               2,
		DefaultSearchText: sqlNullString("", false),
		SelfHostedSyncServerURL: sqlNullString("", false),
	}

	assert.False(t, model.DefaultSearchText.Valid)
	assert.False(t, model.SelfHostedSyncServerURL.Valid)
}
