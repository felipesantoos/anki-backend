package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// DeletionLogRepository implements IDeletionLogRepository using PostgreSQL
type DeletionLogRepository struct {
	db *sql.DB
}

// NewDeletionLogRepository creates a new DeletionLogRepository instance
func NewDeletionLogRepository(db *sql.DB) secondary.IDeletionLogRepository {
	return &DeletionLogRepository{
		db: db,
	}
}

// Save saves or updates a deletion log in the database
func (r *DeletionLogRepository) Save(ctx context.Context, userID int64, deletionLogEntity *deletionlog.DeletionLog) error {
	model := mappers.DeletionLogToModel(deletionLogEntity)

	if deletionLogEntity.GetID() == 0 {
		// Insert new deletion log
		query := `
			INSERT INTO deletions_log (user_id, object_type, object_id, object_data, deleted_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		now := time.Now()
		if model.DeletedAt.IsZero() {
			model.DeletedAt = now
		}

		var deletionLogID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.ObjectType,
			model.ObjectID,
			model.ObjectData,
			model.DeletedAt,
		).Scan(&deletionLogID)
		if err != nil {
			return fmt.Errorf("failed to create deletion log: %w", err)
		}

		deletionLogEntity.SetID(deletionLogID)
		return nil
	}

	// Update existing deletion log - validate ownership first
	existingDeletionLog, err := r.FindByID(ctx, userID, deletionLogEntity.GetID())
	if err != nil {
		return err
	}
	if existingDeletionLog == nil {
		return ownership.ErrResourceNotFound
	}

	// Update deletion log
	query := `
		UPDATE deletions_log
		SET object_type = $1, object_id = $2, object_data = $3, deleted_at = $4
		WHERE id = $5 AND user_id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		model.ObjectType,
		model.ObjectID,
		model.ObjectData,
		model.DeletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update deletion log: %w", err)
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

// FindByID finds a deletion log by ID, filtering by userID to ensure ownership
func (r *DeletionLogRepository) FindByID(ctx context.Context, userID int64, id int64) (*deletionlog.DeletionLog, error) {
	query := `
		SELECT id, user_id, object_type, object_id, object_data, deleted_at
		FROM deletions_log
		WHERE id = $1 AND user_id = $2
	`

	var model models.DeletionLogModel
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.ObjectType,
		&model.ObjectID,
		&model.ObjectData,
		&model.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find deletion log: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.DeletionLogToDomain(&model)
}

// FindByUserID finds all deletion logs for a user
func (r *DeletionLogRepository) FindByUserID(ctx context.Context, userID int64) ([]*deletionlog.DeletionLog, error) {
	query := `
		SELECT id, user_id, object_type, object_id, object_data, deleted_at
		FROM deletions_log
		WHERE user_id = $1
		ORDER BY deleted_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find deletion logs by user ID: %w", err)
	}
	defer rows.Close()

	var deletionLogs []*deletionlog.DeletionLog
	for rows.Next() {
		var model models.DeletionLogModel

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.ObjectType,
			&model.ObjectID,
			&model.ObjectData,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deletion log: %w", err)
		}

		deletionLogEntity, err := mappers.DeletionLogToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert deletion log to domain: %w", err)
		}
		deletionLogs = append(deletionLogs, deletionLogEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deletion logs: %w", err)
	}

	return deletionLogs, nil
}

// Update updates an existing deletion log, validating ownership
func (r *DeletionLogRepository) Update(ctx context.Context, userID int64, id int64, deletionLogEntity *deletionlog.DeletionLog) error {
	return r.Save(ctx, userID, deletionLogEntity)
}

// Delete deletes a deletion log, validating ownership
func (r *DeletionLogRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingDeletionLog, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingDeletionLog == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (deletions_log doesn't have soft delete)
	query := `DELETE FROM deletions_log WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete deletion log: %w", err)
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

// Exists checks if a deletion log exists and belongs to the user
func (r *DeletionLogRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM deletions_log
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check deletion log existence: %w", err)
	}

	return exists, nil
}

// Ensure DeletionLogRepository implements IDeletionLogRepository
var _ secondary.IDeletionLogRepository = (*DeletionLogRepository)(nil)

