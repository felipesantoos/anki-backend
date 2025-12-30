package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	addon "github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestAddOnToDomain_WithAllFields(t *testing.T) {
	now := time.Now()

	model := &models.AddOnModel{
		ID:          1,
		UserID:      100,
		Code:        "1234567890",
		Name:        "Awesome Add-on",
		Version:     "1.0.0",
		Enabled:     true,
		ConfigJSON:  `{"setting1":"value1"}`,
		InstalledAt: now,
		UpdatedAt:   now.Add(time.Hour),
	}

	entity, err := AddOnToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "1234567890", entity.GetCode())
	assert.Equal(t, "Awesome Add-on", entity.GetName())
	assert.Equal(t, "1.0.0", entity.GetVersion())
	assert.True(t, entity.GetEnabled())
	assert.Equal(t, `{"setting1":"value1"}`, entity.GetConfigJSON())
	assert.Equal(t, now, entity.GetInstalledAt())
	assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
}

func TestAddOnToDomain_Disabled(t *testing.T) {
	now := time.Now()

	model := &models.AddOnModel{
		ID:          2,
		UserID:      200,
		Code:        "0987654321",
		Name:        "Disabled Add-on",
		Version:     "2.0.0",
		Enabled:     false,
		ConfigJSON:  `{}`,
		InstalledAt: now,
		UpdatedAt:   now,
	}

	entity, err := AddOnToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.False(t, entity.GetEnabled())
}

func TestAddOnToDomain_NilInput(t *testing.T) {
	entity, err := AddOnToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestAddOnToModel_WithAllFields(t *testing.T) {
	now := time.Now()

	entity, err := addon.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithCode("1234567890").
		WithName("Awesome Add-on").
		WithVersion("1.0.0").
		WithEnabled(true).
		WithConfigJSON(`{"setting1":"value1"}`).
		WithInstalledAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		Build()
	require.NoError(t, err)

	model := AddOnToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, int64(100), model.UserID)
	assert.Equal(t, "1234567890", model.Code)
	assert.Equal(t, "Awesome Add-on", model.Name)
	assert.Equal(t, "1.0.0", model.Version)
	assert.True(t, model.Enabled)
	assert.Equal(t, `{"setting1":"value1"}`, model.ConfigJSON)
	assert.Equal(t, now, model.InstalledAt)
	assert.Equal(t, now.Add(time.Hour), model.UpdatedAt)
}

