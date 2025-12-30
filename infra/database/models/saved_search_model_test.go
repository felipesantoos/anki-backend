package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSavedSearchModel_Creation(t *testing.T) {
	now := time.Now()
	model := &SavedSearchModel{
		ID:          1,
		UserID:      100,
		Name:        "Due Cards",
		SearchQuery: "is:due",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Due Cards", model.Name)
}

