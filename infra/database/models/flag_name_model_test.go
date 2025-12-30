package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlagNameModel_Creation(t *testing.T) {
	now := time.Now()
	model := &FlagNameModel{
		ID:         1,
		UserID:     100,
		FlagNumber: 1,
		Name:       "Important",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, 1, model.FlagNumber)
}

