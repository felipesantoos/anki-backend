package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
)

// ICheckDatabaseLogService defines the interface for database integrity check logs
type ICheckDatabaseLogService interface {
	// Create records a new database check result
	Create(ctx context.Context, userID int64, result string, errorsFound int) (*checkdatabaselog.CheckDatabaseLog, error)

	// FindLatest finds the most recent database check logs for a user
	FindLatest(ctx context.Context, userID int64, limit int) ([]*checkdatabaselog.CheckDatabaseLog, error)
}

