package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	flagname "github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// FlagNameRepository implements IFlagNameRepository using PostgreSQL
type FlagNameRepository struct {
	db *sql.DB
}

// NewFlagNameRepository creates a new FlagNameRepository instance
func NewFlagNameRepository(db *sql.DB) secondary.IFlagNameRepository {
	return &FlagNameRepository{
		db: db,
	}
}

// Save saves or updates a flag name in the database
func (r *FlagNameRepository) Save(ctx context.Context, userID int64, flagNameEntity *flagname.FlagName) error {
	model := mappers.FlagNameToModel(flagNameEntity)

	if flagNameEntity.GetID() == 0 {
		// Insert new flag name
		query := `
			INSERT INTO flag_names (user_id, flag_number, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var flagNameID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.FlagNumber,
			model.Name,
			model.CreatedAt,
			model.UpdatedAt,
		).Scan(&flagNameID)
		if err != nil {
			return fmt.Errorf("failed to create flag name: %w", err)
		}

		flagNameEntity.SetID(flagNameID)
		return nil
	}

	// Update existing flag name - validate ownership first
	existingFlagName, err := r.FindByID(ctx, userID, flagNameEntity.GetID())
	if err != nil {
		return err
	}
	if existingFlagName == nil {
		return ownership.ErrResourceNotFound
	}

	// Update flag name
	query := `
		UPDATE flag_names
		SET flag_number = $1, name = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5
	`

	now := time.Now()
	model.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		model.FlagNumber,
		model.Name,
		model.UpdatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update flag name: %w", err)
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

// FindByID finds a flag name by ID, filtering by userID to ensure ownership
func (r *FlagNameRepository) FindByID(ctx context.Context, userID int64, id int64) (*flagname.FlagName, error) {
	query := `
		SELECT id, user_id, flag_number, name, created_at, updated_at
		FROM flag_names
		WHERE id = $1 AND user_id = $2
	`

	var model models.FlagNameModel
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.FlagNumber,
		&model.Name,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find flag name: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.FlagNameToDomain(&model)
}

// FindByUserID finds all flag names for a user
func (r *FlagNameRepository) FindByUserID(ctx context.Context, userID int64) ([]*flagname.FlagName, error) {
	query := `
		SELECT id, user_id, flag_number, name, created_at, updated_at
		FROM flag_names
		WHERE user_id = $1
		ORDER BY flag_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find flag names by user ID: %w", err)
	}
	defer rows.Close()

	var flagNames []*flagname.FlagName
	for rows.Next() {
		var model models.FlagNameModel

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.FlagNumber,
			&model.Name,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag name: %w", err)
		}

		flagNameEntity, err := mappers.FlagNameToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert flag name to domain: %w", err)
		}
		flagNames = append(flagNames, flagNameEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flag names: %w", err)
	}

	return flagNames, nil
}

// Update updates an existing flag name, validating ownership
func (r *FlagNameRepository) Update(ctx context.Context, userID int64, id int64, flagNameEntity *flagname.FlagName) error {
	return r.Save(ctx, userID, flagNameEntity)
}

// Delete deletes a flag name, validating ownership
func (r *FlagNameRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingFlagName, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingFlagName == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (flag_names don't have soft delete)
	query := `DELETE FROM flag_names WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete flag name: %w", err)
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

// Exists checks if a flag name exists and belongs to the user
func (r *FlagNameRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM flag_names
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check flag name existence: %w", err)
	}

	return exists, nil
}

// FindByFlagNumber finds a flag name by flag number, filtering by userID to ensure ownership
func (r *FlagNameRepository) FindByFlagNumber(ctx context.Context, userID int64, flagNumber int) (*flagname.FlagName, error) {
	query := `
		SELECT id, user_id, flag_number, name, created_at, updated_at
		FROM flag_names
		WHERE flag_number = $1 AND user_id = $2
	`

	var model models.FlagNameModel
	err := r.db.QueryRowContext(ctx, query, flagNumber, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.FlagNumber,
		&model.Name,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find flag name by number: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.FlagNameToDomain(&model)
}

// Ensure FlagNameRepository implements IFlagNameRepository
var _ secondary.IFlagNameRepository = (*FlagNameRepository)(nil)

