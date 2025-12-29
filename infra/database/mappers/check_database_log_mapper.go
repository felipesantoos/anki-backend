package mappers

import (
	"database/sql"

	checkdatabaselog "github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// CheckDatabaseLogToDomain converts a CheckDatabaseLogModel (database representation) to a CheckDatabaseLog entity (domain representation)
func CheckDatabaseLogToDomain(model *models.CheckDatabaseLogModel) (*checkdatabaselog.CheckDatabaseLog, error) {
	if model == nil {
		return nil, nil
	}

	builder := checkdatabaselog.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithStatus(model.Status).
		WithIssuesFound(model.IssuesFound).
		WithIssuesDetails(model.IssuesDetails).
		WithCreatedAt(model.CreatedAt)

	// Handle nullable execution_time_ms (convert int64 to int)
	if model.ExecutionTimeMs.Valid {
		execTimeMs := int(model.ExecutionTimeMs.Int64)
		builder = builder.WithExecutionTimeMs(&execTimeMs)
	}

	return builder.Build()
}

// CheckDatabaseLogToModel converts a CheckDatabaseLog entity (domain representation) to a CheckDatabaseLogModel (database representation)
func CheckDatabaseLogToModel(checkDatabaseLogEntity *checkdatabaselog.CheckDatabaseLog) *models.CheckDatabaseLogModel {
	model := &models.CheckDatabaseLogModel{
		ID:            checkDatabaseLogEntity.GetID(),
		UserID:        checkDatabaseLogEntity.GetUserID(),
		Status:        checkDatabaseLogEntity.GetStatus(),
		IssuesFound:   checkDatabaseLogEntity.GetIssuesFound(),
		IssuesDetails: checkDatabaseLogEntity.GetIssuesDetails(),
		CreatedAt:     checkDatabaseLogEntity.GetCreatedAt(),
	}

	// Handle nullable execution_time_ms (convert int to int64)
	if checkDatabaseLogEntity.GetExecutionTimeMs() != nil {
		execTimeMs := int64(*checkDatabaseLogEntity.GetExecutionTimeMs())
		model.ExecutionTimeMs = sql.NullInt64{
			Int64: execTimeMs,
			Valid: true,
		}
	}

	return model
}

