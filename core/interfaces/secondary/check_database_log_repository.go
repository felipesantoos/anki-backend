package secondary

import (
	"context"

	checkdatabaselog "github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
)

// ICheckDatabaseLogRepository defines the interface for check database log data persistence
// All methods that access specific resources require userID to ensure data isolation
type ICheckDatabaseLogRepository interface {
	// Save saves or updates a check database log entry in the database
	// If the check database log has an ID, it updates the existing entry
	// If the check database log has no ID, it creates a new entry and returns it with the ID set
	Save(ctx context.Context, userID int64, checkDatabaseLogEntity *checkdatabaselog.CheckDatabaseLog) error

	// FindByID finds a check database log entry by ID, filtering by userID to ensure ownership
	// Returns the check database log if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*checkdatabaselog.CheckDatabaseLog, error)

	// FindByUserID finds all check database log entries for a user
	// Returns a list of check database log entries belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*checkdatabaselog.CheckDatabaseLog, error)

	// Update updates an existing check database log entry, validating ownership
	// Returns error if check database log doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, checkDatabaseLogEntity *checkdatabaselog.CheckDatabaseLog) error

	// Delete deletes a check database log entry, validating ownership (hard delete - check_database_log doesn't have soft delete)
	// Returns error if check database log doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a check database log entry exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindLatest finds the latest check logs for a user
	FindLatest(ctx context.Context, userID int64, limit int) ([]*checkdatabaselog.CheckDatabaseLog, error)
}

