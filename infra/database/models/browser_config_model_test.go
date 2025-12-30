package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBrowserConfigModel_Creation(t *testing.T) {
	now := time.Now()
	model := &BrowserConfigModel{
		ID:            1,
		UserID:        100,
		VisibleColumns: "{id,front,back}",
		ColumnWidths:   `{"id":100}`,
		SortColumn:     "due",
		SortDirection:  "asc",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "due", model.SortColumn)
}

