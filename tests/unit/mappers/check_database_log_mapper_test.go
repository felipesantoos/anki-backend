package mappers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	checkdatabaselog "github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

func TestCheckDatabaseLogToDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	execTimeMs := int64(1500)

	model := &models.CheckDatabaseLogModel{
		ID:             1,
		UserID:         100,
		Status:         "success",
		IssuesFound:    0,
		IssuesDetails:  `[]`,
		ExecutionTimeMs: sqlNullInt64(execTimeMs, true),
		CreatedAt:      now,
	}

	entity, err := CheckDatabaseLogToDomain(model)
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.Equal(t, int64(1), entity.GetID())
	assert.Equal(t, int64(100), entity.GetUserID())
	assert.Equal(t, "success", entity.GetStatus())
	assert.Equal(t, 0, entity.GetIssuesFound())
	assert.NotNil(t, entity.GetExecutionTimeMs())
	assert.Equal(t, 1500, *entity.GetExecutionTimeMs())
}

func TestCheckDatabaseLogToDomain_WithNullExecutionTime(t *testing.T) {
	now := time.Now()

	model := &models.CheckDatabaseLogModel{
		ID:             2,
		UserID:         200,
		Status:         "error",
		IssuesFound:    5,
		IssuesDetails:  `[{"issue":"missing note"}]`,
		ExecutionTimeMs: sqlNullInt64(0, false),
		CreatedAt:      now,
	}

	entity, err := CheckDatabaseLogToDomain(model)
	require.NoError(t, err)
	assert.Nil(t, entity.GetExecutionTimeMs())
}

func TestCheckDatabaseLogToDomain_NilInput(t *testing.T) {
	entity, err := CheckDatabaseLogToDomain(nil)
	assert.NoError(t, err)
	assert.Nil(t, entity)
}

func TestCheckDatabaseLogToModel_WithAllFields(t *testing.T) {
	now := time.Now()
	execTimeMs := 1500

	entity, err := checkdatabaselog.NewBuilder().
		WithID(1).
		WithUserID(100).
		WithStatus("success").
		WithIssuesFound(0).
		WithIssuesDetails(`[]`).
		WithExecutionTimeMs(&execTimeMs).
		WithCreatedAt(now).
		Build()
	require.NoError(t, err)

	model := CheckDatabaseLogToModel(entity)
	require.NotNil(t, model)

	assert.Equal(t, int64(1), model.ID)
	assert.True(t, model.ExecutionTimeMs.Valid)
	assert.Equal(t, int64(1500), model.ExecutionTimeMs.Int64)
}

