package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	browserconfig "github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestBrowserConfigToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	sortColumn := "due"

	model := &models.BrowserConfigModel{
		ID:            1,
		UserID:        100,
		VisibleColumns: "{id,front,back}",
		ColumnWidths:  `{"id":100,"front":200}`,
		SortColumn:     sortColumn,
		SortDirection:  "asc",
		CreatedAt:      now,
		UpdatedAt:      now.Add(time.Hour),
	}

	entity, err := BrowserConfigToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, []string{"id", "front", "back"}, entity.GetVisibleColumns())
	assert.Equal(t, `{"id":100,"front":200}`, entity.GetColumnWidths())
	assert.NotNil(t, entity.GetSortColumn())
	assert.Equal(t, sortColumn, *entity.GetSortColumn())
	assert.Equal(t, "asc", entity.GetSortDirection())
}

func TestBrowserConfigToDomain_WithEmptySortColumn(t *testing.T) {
	now := time.Now()

	model := &models.BrowserConfigModel{
		ID:            1,
		UserID:        100,
		VisibleColumns: "{}",
		ColumnWidths:  `{}`,
		SortColumn:     "",
		SortDirection:  "asc",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	entity, err := BrowserConfigToDomain(model)
	require.NoError(t, err)
	assert.Nil(t, entity.GetSortColumn())
}

func TestBrowserConfigToDomain_NilInput(t *testing.T) {
	entity, err := BrowserConfigToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestBrowserConfigToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	sortColumn := "due"

	entity, err := browserconfig.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithVisibleColumns([]string{"id", "front"}).
		WithColumnWidths(`{"id":100}`).
		WithSortColumn(&sortColumn).
		WithSortDirection("asc").
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		Build()
	require.NoError(t, err)

	model := BrowserConfigToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, sortColumn, model.SortColumn)
}

