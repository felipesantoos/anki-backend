package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	filtereddeck "github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestFilteredDeckToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	lastRebuildAt := now.Add(-time.Hour)
	deletedAt := now.Add(time.Hour)
	secondFilter := "deck:Default"

	model := &models.FilteredDeckModel{
		ID:           1,
		UserID:       100,
		Name:         "Due Cards",
		SearchFilter: "is:due",
		SecondFilter: sqlNullString(secondFilter, true),
		LimitCards:   20,
		OrderBy:      "due",
		Reschedule:   true,
		CreatedAt:    now,
		UpdatedAt:   now.Add(time.Hour),
		LastRebuildAt: sqlNullTime(lastRebuildAt, true),
		DeletedAt:    sqlNullTime(deletedAt, true),
	}

	entity, err := FilteredDeckToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "Due Cards", entity.GetName())
	assert.Equal(t, "is:due", entity.GetSearchFilter())
	assert.NotNil(t, entity.GetSecondFilter())
	assert.Equal(t, secondFilter, *entity.GetSecondFilter())
	assert.Equal(t, 20, entity.GetLimitCards())
	assert.Equal(t, "due", entity.GetOrderBy())
	assert.True(t, entity.GetReschedule())
	assert.Equal(t, now, entity.GetCreatedAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
	assert.NotNil(t, entity.GetLastRebuildAt())
	assert.Equal(t, lastRebuildAt, *entity.GetLastRebuildAt())
	assert.NotNil(t, entity.GetDeletedAt())
	assert.Equal(t, deletedAt, *entity.GetDeletedAt())
}

func TestFilteredDeckToDomain_WithNullFields(t *testing.T) {
	now := time.Now()

	model := &models.FilteredDeckModel{
		ID:           2,
		UserID:       200,
		Name:         "New Cards",
		SearchFilter: "is:new",
		SecondFilter: sqlNullString("", false),
		LimitCards:   10,
		OrderBy:      "created",
		Reschedule:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRebuildAt: sqlNullTime(time.Time{}, false),
		DeletedAt:    sqlNullTime(time.Time{}, false),
	}

	entity, err := FilteredDeckToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Nil(t, entity.GetSecondFilter())
	assert.Nil(t, entity.GetLastRebuildAt())
	assert.Nil(t, entity.GetDeletedAt())
}

func TestFilteredDeckToDomain_NilInput(t *testing.T) {
	entity, err := FilteredDeckToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestFilteredDeckToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	lastRebuildAt := now.Add(-time.Hour)
	deletedAt := now.Add(time.Hour)
	secondFilter := "deck:Default"

	entity, err := filtereddeck.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithName("Due Cards").
		WithSearchFilter("is:due").
		WithSecondFilter(&secondFilter).
		WithLimitCards(20).
		WithOrderBy("due").
		WithReschedule(true).
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		WithLastRebuildAt(&lastRebuildAt).
		WithDeletedAt(&deletedAt).
		Build()
	require.NoError(t, err)

	model := FilteredDeckToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "Due Cards", model.Name)
	assert.Equal(t, "is:due", model.SearchFilter)
	assert.True(t, model.SecondFilter.Valid)
	assert.Equal(t, secondFilter, model.SecondFilter.String)
	assert.Equal(t, 20, model.LimitCards)
	assert.Equal(t, "due", model.OrderBy)
	assert.True(t, model.Reschedule)
	assert.Equal(t, now, model.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
	assert.True(t, model.LastRebuildAt.Valid)
	assert.Equal(t, lastRebuildAt, model.LastRebuildAt.Time)
	assert.True(t, model.DeletedAt.Valid)
	assert.Equal(t, deletedAt, model.DeletedAt.Time)
}

func TestFilteredDeckToModel_WithNullFields(t *testing.T) {
	now := time.Now()

	entity, err := filtereddeck.NewBuilder().
		WithID(2).
		WithUserID(200).
		WithName("New Cards").
		WithSearchFilter("is:new").
		WithSecondFilter(nil).
		WithLimitCards(10).
		WithOrderBy("created").
		WithReschedule(false).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		WithLastRebuildAt(nil).
		WithDeletedAt(nil).
		Build()
	require.NoError(t, err)

	model := FilteredDeckToModel(entity)
	require.NotNil(t, model)

	assert.False(t, model.SecondFilter.Valid)
	assert.False(t, model.LastRebuildAt.Valid)
	assert.False(t, model.DeletedAt.Valid)
}

