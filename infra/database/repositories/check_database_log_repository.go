package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	checkdatabaselog "github.com/felipesantos/anki-backend/core/domain/entities/check_database_log"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// CheckDatabaseLogRepository implements ICheckDatabaseLogRepository using PostgreSQL
type CheckDatabaseLogRepository struct {
	db *sql.DB
}

// NewCheckDatabaseLogRepository creates a new CheckDatabaseLogRepository instance
func NewCheckDatabaseLogRepository(db *sql.DB) secondary.ICheckDatabaseLogRepository {
	return &CheckDatabaseLogRepository{
		db: db,
	}
}

// Save saves or updates a check database log entry in the database
func (r *CheckDatabaseLogRepository) Save(ctx context.Context, userID int64, checkDatabaseLogEntity *checkdatabaselog.CheckDatabaseLog) error {
	model := mappers.CheckDatabaseLogToModel(checkDatabaseLogEntity)

	if checkDatabaseLogEntity.GetID() == 0 {
		// Insert new check database log entry
		query := `
			INSERT INTO check_database_log (user_id, status, issues_found, issues_details, execution_time_ms, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}

		var executionTimeMs interface{}
		if model.ExecutionTimeMs.Valid {
			executionTimeMs = model.ExecutionTimeMs.Int64
		}

		var checkDatabaseLogID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Status,
			model.IssuesFound,
			model.IssuesDetails,
			executionTimeMs,
			model.CreatedAt,
		).Scan(&checkDatabaseLogID)
		if err != nil {
			return fmt.Errorf("failed to create check database log: %w", err)
		}

		checkDatabaseLogEntity.SetID(checkDatabaseLogID)
		return nil
	}

	// Update existing check database log entry - validate ownership first
	existingCheckDatabaseLog, err := r.FindByID(ctx, userID, checkDatabaseLogEntity.GetID())
	if err != nil {
		return err
	}
	if existingCheckDatabaseLog == nil {
		return ownership.ErrResourceNotFound
	}

	// Update check database log entry
	query := `
		UPDATE check_database_log
		SET status = $1, issues_found = $2, issues_details = $3, execution_time_ms = $4, created_at = $5
		WHERE id = $6 AND user_id = $7
	`

	var executionTimeMs interface{}
	if model.ExecutionTimeMs.Valid {
		executionTimeMs = model.ExecutionTimeMs.Int64
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Status,
		model.IssuesFound,
		model.IssuesDetails,
		executionTimeMs,
		model.CreatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update check database log: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// FindByID finds a check database log entry by ID, filtering by userID to ensure ownership
func (r *CheckDatabaseLogRepository) FindByID(ctx context.Context, userID int64, id int64) (*checkdatabaselog.CheckDatabaseLog, error) {
	query := `
		SELECT id, user_id, status, issues_found, issues_details, execution_time_ms, created_at
		FROM check_database_log
		WHERE id = $1 AND user_id = $2
	`

	var model models.CheckDatabaseLogModel
	var executionTimeMs sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Status,
		&model.IssuesFound,
		&model.IssuesDetails,
		&executionTimeMs,
		&model.CreatedAt,
	)

	model.ExecutionTimeMs = executionTimeMs

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find check database log: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.CheckDatabaseLogToDomain(&model)
}

// FindByUserID finds all check database log entries for a user
func (r *CheckDatabaseLogRepository) FindByUserID(ctx context.Context, userID int64) ([]*checkdatabaselog.CheckDatabaseLog, error) {
	query := `
		SELECT id, user_id, status, issues_found, issues_details, execution_time_ms, created_at
		FROM check_database_log
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find check database logs by user ID: %w", err)
	}
	defer rows.Close()

	var checkDatabaseLogs []*checkdatabaselog.CheckDatabaseLog
	for rows.Next() {
		var model models.CheckDatabaseLogModel
		var executionTimeMs sql.NullInt64

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Status,
			&model.IssuesFound,
			&model.IssuesDetails,
			&executionTimeMs,
			&model.CreatedAt,
		)

		model.ExecutionTimeMs = executionTimeMs
		if err != nil {
			return nil, fmt.Errorf("failed to scan check database log: %w", err)
		}

		checkDatabaseLogEntity, err := mappers.CheckDatabaseLogToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert check database log to domain: %w", err)
		}
		checkDatabaseLogs = append(checkDatabaseLogs, checkDatabaseLogEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating check database logs: %w", err)
	}

	return checkDatabaseLogs, nil
}

// Update updates an existing check database log entry, validating ownership
func (r *CheckDatabaseLogRepository) Update(ctx context.Context, userID int64, id int64, checkDatabaseLogEntity *checkdatabaselog.CheckDatabaseLog) error {
	return r.Save(ctx, userID, checkDatabaseLogEntity)
}

// Delete deletes a check database log entry, validating ownership
func (r *CheckDatabaseLogRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingCheckDatabaseLog, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingCheckDatabaseLog == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (check_database_log doesn't have soft delete)
	query := `DELETE FROM check_database_log WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete check database log: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// Exists checks if a check database log entry exists and belongs to the user
func (r *CheckDatabaseLogRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM check_database_log
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check check database log existence: %w", err)
	}

	return exists, nil
}

// Ensure CheckDatabaseLogRepository implements ICheckDatabaseLogRepository
var _ secondary.ICheckDatabaseLogRepository = (*CheckDatabaseLogRepository)(nil)

