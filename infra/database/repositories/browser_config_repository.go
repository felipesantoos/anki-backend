package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	browserconfig "github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// BrowserConfigRepository implements IBrowserConfigRepository using PostgreSQL
type BrowserConfigRepository struct {
	db *sql.DB
}

// NewBrowserConfigRepository creates a new BrowserConfigRepository instance
func NewBrowserConfigRepository(db *sql.DB) secondary.IBrowserConfigRepository {
	return &BrowserConfigRepository{
		db: db,
	}
}

// Save saves or updates browser config in the database
func (r *BrowserConfigRepository) Save(ctx context.Context, userID int64, browserConfigEntity *browserconfig.BrowserConfig) error {
	model := mappers.BrowserConfigToModel(browserConfigEntity)

	if browserConfigEntity.GetID() == 0 {
		// Insert new config
		query := `
			INSERT INTO browser_config (user_id, visible_columns, column_widths, sort_column, sort_direction, created_at, updated_at)
			VALUES ($1, $2::TEXT[], $3, $4, $5, $6, $7)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		// Get visible columns from entity
		visibleColumns := browserConfigEntity.GetVisibleColumns()

		var configID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			pq.Array(visibleColumns),
			model.ColumnWidths,
			model.SortColumn,
			model.SortDirection,
			model.CreatedAt,
			model.UpdatedAt,
		).Scan(&configID)
		if err != nil {
			return fmt.Errorf("failed to create browser config: %w", err)
		}

		browserConfigEntity.SetID(configID)
		return nil
	}

	// Update existing config - validate ownership first
	existingConfig, err := r.FindByID(ctx, userID, browserConfigEntity.GetID())
	if err != nil {
		return err
	}
	if existingConfig == nil {
		return ownership.ErrResourceNotFound
	}

	// Update config
	query := `
		UPDATE browser_config
		SET visible_columns = $1::TEXT[], column_widths = $2, sort_column = $3, sort_direction = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7
	`

	now := time.Now()
	model.UpdatedAt = now

	// Get visible columns from entity
	visibleColumns := browserConfigEntity.GetVisibleColumns()

	result, err := r.db.ExecContext(ctx, query,
		pq.Array(visibleColumns),
		model.ColumnWidths,
		model.SortColumn,
		model.SortDirection,
		model.UpdatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update browser config: %w", err)
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

// FindByID finds browser config by ID, filtering by userID to ensure ownership
func (r *BrowserConfigRepository) FindByID(ctx context.Context, userID int64, id int64) (*browserconfig.BrowserConfig, error) {
	query := `
		SELECT id, user_id, visible_columns, column_widths, sort_column, sort_direction, created_at, updated_at
		FROM browser_config
		WHERE id = $1 AND user_id = $2
	`

	var model models.BrowserConfigModel
	var visibleColumns pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&visibleColumns,
		&model.ColumnWidths,
		&model.SortColumn,
		&model.SortDirection,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find browser config: %w", err)
	}

	// Convert pq.StringArray to string format for model (comma-separated within braces)
	visibleColumnsStr := "{"
	for i, col := range visibleColumns {
		if i > 0 {
			visibleColumnsStr += ","
		}
		visibleColumnsStr += col
	}
	visibleColumnsStr += "}"
	model.VisibleColumns = visibleColumnsStr

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.BrowserConfigToDomain(&model)
}

// FindByUserID finds browser config for a user (one-to-one relationship)
func (r *BrowserConfigRepository) FindByUserID(ctx context.Context, userID int64) (*browserconfig.BrowserConfig, error) {
	query := `
		SELECT id, user_id, visible_columns, column_widths, sort_column, sort_direction, created_at, updated_at
		FROM browser_config
		WHERE user_id = $1
	`

	var model models.BrowserConfigModel
	var visibleColumns pq.StringArray

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&model.ID,
		&model.UserID,
		&visibleColumns,
		&model.ColumnWidths,
		&model.SortColumn,
		&model.SortDirection,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find browser config by user ID: %w", err)
	}

	// Convert pq.StringArray to string format for model (comma-separated within braces)
	visibleColumnsStr := "{"
	for i, col := range visibleColumns {
		if i > 0 {
			visibleColumnsStr += ","
		}
		visibleColumnsStr += col
	}
	visibleColumnsStr += "}"
	model.VisibleColumns = visibleColumnsStr

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.BrowserConfigToDomain(&model)
}

// Update updates existing browser config, validating ownership
func (r *BrowserConfigRepository) Update(ctx context.Context, userID int64, id int64, browserConfigEntity *browserconfig.BrowserConfig) error {
	return r.Save(ctx, userID, browserConfigEntity)
}

// Delete deletes browser config, validating ownership
func (r *BrowserConfigRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingConfig, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingConfig == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (browser_config doesn't have soft delete)
	query := `DELETE FROM browser_config WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete browser config: %w", err)
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

// Exists checks if browser config exists for a user
func (r *BrowserConfigRepository) Exists(ctx context.Context, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM browser_config
			WHERE user_id = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check browser config existence: %w", err)
	}

	return exists, nil
}

// Ensure BrowserConfigRepository implements IBrowserConfigRepository
var _ secondary.IBrowserConfigRepository = (*BrowserConfigRepository)(nil)

