package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	savedsearch "github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestSavedSearchToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	model := &models.SavedSearchModel{
		ID:          1,
		UserID:      100,
		Name:        "Due Cards",
		SearchQuery: "is:due",
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Hour),
		DeletedAt:   sqlNullTime(deletedAt, true),
	}

	entity, err := SavedSearchToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, "Due Cards", entity.GetName())
	assert.Equal(t, "is:due", entity.GetSearchQuery())
	assert.NotNil(t, entity.GetDeletedAt())
}

func TestSavedSearchToDomain_NilInput(t *testing.T) {
	entity, err := SavedSearchToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestSavedSearchToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	entity, err := savedsearch.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithName("Due Cards").
		WithSearchQuery("is:due").
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		WithDeletedAt(&deletedAt).
		Build()
	require.NoError(t, err)

	model := SavedSearchToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "Due Cards", model.Name)
	assert.True(t, model.DeletedAt.Valid)
}

