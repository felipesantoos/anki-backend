package mappers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flagname "github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestFlagNameToDomain_WithAllFields(t *testing.T) {
	now := time.Now()

	model := &models.FlagNameModel{
		ID:         1,
		UserID:     100,
		FlagNumber: 1,
		Name:       "Important",
		CreatedAt:  now,
		UpdatedAt:  now.Add(time.Hour),
	}

	entity, err := FlagNameToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, 1, entity.GetFlagNumber())
	assert.Equal(t, "Important", entity.GetName())
}

func TestFlagNameToDomain_NilInput(t *testing.T) {
	entity, err := FlagNameToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestFlagNameToModel_WithAllFields(t *testing.T) {
	now := time.Now()

	entity, err := flagname.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithFlagNumber(1).
		WithName("Important").
		WithCreatedAt(now).
		WithUpdatedAt(now.Add(time.Hour)).
		Build()
	require.NoError(t, err)

	model := FlagNameToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, 1, model.FlagNumber)
	assert.Equal(t, "Important", model.Name)
}

