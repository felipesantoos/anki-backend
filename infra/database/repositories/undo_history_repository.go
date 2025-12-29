package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// UndoHistoryRepository implements IUndoHistoryRepository using PostgreSQL
type UndoHistoryRepository struct {
	db *sql.DB
}

// NewUndoHistoryRepository creates a new UndoHistoryRepository instance
func NewUndoHistoryRepository(db *sql.DB) secondary.IUndoHistoryRepository {
	return &UndoHistoryRepository{
		db: db,
	}
}

// Save saves or updates an undo history entry in the database
func (r *UndoHistoryRepository) Save(ctx context.Context, userID int64, undoHistoryEntity *undohistory.UndoHistory) error {
	model := mappers.UndoHistoryToModel(undoHistoryEntity)

	if undoHistoryEntity.GetID() == 0 {
		// Insert new undo history entry
		query := `
			INSERT INTO undo_history (user_id, operation_type, operation_data, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}

		var undoHistoryID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.OperationType,
			model.OperationData,
			model.CreatedAt,
		).Scan(&undoHistoryID)
		if err != nil {
			return fmt.Errorf("failed to create undo history: %w", err)
		}

		undoHistoryEntity.SetID(undoHistoryID)
		return nil
	}

	// Update existing undo history entry - validate ownership first
	existingUndoHistory, err := r.FindByID(ctx, userID, undoHistoryEntity.GetID())
	if err != nil {
		return err
	}
	if existingUndoHistory == nil {
		return ownership.ErrResourceNotFound
	}

	// Update undo history entry
	query := `
		UPDATE undo_history
		SET operation_type = $1, operation_data = $2, created_at = $3
		WHERE id = $4 AND user_id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		model.OperationType,
		model.OperationData,
		model.CreatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update undo history: %w", err)
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

// FindByID finds an undo history entry by ID, filtering by userID to ensure ownership
func (r *UndoHistoryRepository) FindByID(ctx context.Context, userID int64, id int64) (*undohistory.UndoHistory, error) {
	query := `
		SELECT id, user_id, operation_type, operation_data, created_at
		FROM undo_history
		WHERE id = $1 AND user_id = $2
	`

	var model models.UndoHistoryModel
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.OperationType,
		&model.OperationData,
		&model.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find undo history: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.UndoHistoryToDomain(&model)
}

// FindByUserID finds all undo history entries for a user
func (r *UndoHistoryRepository) FindByUserID(ctx context.Context, userID int64) ([]*undohistory.UndoHistory, error) {
	query := `
		SELECT id, user_id, operation_type, operation_data, created_at
		FROM undo_history
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find undo history by user ID: %w", err)
	}
	defer rows.Close()

	var undoHistories []*undohistory.UndoHistory
	for rows.Next() {
		var model models.UndoHistoryModel

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.OperationType,
			&model.OperationData,
			&model.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan undo history: %w", err)
		}

		undoHistoryEntity, err := mappers.UndoHistoryToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert undo history to domain: %w", err)
		}
		undoHistories = append(undoHistories, undoHistoryEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating undo history: %w", err)
	}

	return undoHistories, nil
}

// Update updates an existing undo history entry, validating ownership
func (r *UndoHistoryRepository) Update(ctx context.Context, userID int64, id int64, undoHistoryEntity *undohistory.UndoHistory) error {
	return r.Save(ctx, userID, undoHistoryEntity)
}

// Delete deletes an undo history entry, validating ownership
func (r *UndoHistoryRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingUndoHistory, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingUndoHistory == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (undo_history doesn't have soft delete)
	query := `DELETE FROM undo_history WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete undo history: %w", err)
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

// Exists checks if an undo history entry exists and belongs to the user
func (r *UndoHistoryRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM undo_history
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check undo history existence: %w", err)
	}

	return exists, nil
}

// Ensure UndoHistoryRepository implements IUndoHistoryRepository
var _ secondary.IUndoHistoryRepository = (*UndoHistoryRepository)(nil)

