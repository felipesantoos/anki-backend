package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	savedsearch "github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// SavedSearchRepository implements ISavedSearchRepository using PostgreSQL
type SavedSearchRepository struct {
	db *sql.DB
}

// NewSavedSearchRepository creates a new SavedSearchRepository instance
func NewSavedSearchRepository(db *sql.DB) secondary.ISavedSearchRepository {
	return &SavedSearchRepository{
		db: db,
	}
}

// Save saves or updates a saved search in the database
func (r *SavedSearchRepository) Save(ctx context.Context, userID int64, savedSearchEntity *savedsearch.SavedSearch) error {
	model := mappers.SavedSearchToModel(savedSearchEntity)

	if savedSearchEntity.GetID() == 0 {
		// Insert new saved search
		query := `
			INSERT INTO saved_searches (user_id, name, search_query, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var deletedAt sql.NullTime
		if model.DeletedAt.Valid {
			deletedAt = model.DeletedAt
		}

		var savedSearchID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Name,
			model.SearchQuery,
			model.CreatedAt,
			model.UpdatedAt,
			deletedAt,
		).Scan(&savedSearchID)
		if err != nil {
			return fmt.Errorf("failed to create saved search: %w", err)
		}

		savedSearchEntity.SetID(savedSearchID)
		return nil
	}

	// Update existing saved search - validate ownership first
	existingSavedSearch, err := r.FindByID(ctx, userID, savedSearchEntity.GetID())
	if err != nil {
		return err
	}
	if existingSavedSearch == nil {
		return ownership.ErrResourceNotFound
	}

	// Update saved search
	query := `
		UPDATE saved_searches
		SET name = $1, search_query = $2, updated_at = $3, deleted_at = $4
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	`

	now := time.Now()
	model.UpdatedAt = now

	var deletedAt sql.NullTime
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Name,
		model.SearchQuery,
		model.UpdatedAt,
		deletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update saved search: %w", err)
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

// FindByID finds a saved search by ID, filtering by userID to ensure ownership
func (r *SavedSearchRepository) FindByID(ctx context.Context, userID int64, id int64) (*savedsearch.SavedSearch, error) {
	query := `
		SELECT id, user_id, name, search_query, created_at, updated_at, deleted_at
		FROM saved_searches
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.SavedSearchModel
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.SearchQuery,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find saved search: %w", err)
	}

	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.SavedSearchToDomain(&model)
}

// FindByUserID finds all saved searches for a user
func (r *SavedSearchRepository) FindByUserID(ctx context.Context, userID int64) ([]*savedsearch.SavedSearch, error) {
	query := `
		SELECT id, user_id, name, search_query, created_at, updated_at, deleted_at
		FROM saved_searches
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find saved searches by user ID: %w", err)
	}
	defer rows.Close()

	var savedSearches []*savedsearch.SavedSearch
	for rows.Next() {
		var model models.SavedSearchModel
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.SearchQuery,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan saved search: %w", err)
		}

		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		savedSearchEntity, err := mappers.SavedSearchToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert saved search to domain: %w", err)
		}
		savedSearches = append(savedSearches, savedSearchEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating saved searches: %w", err)
	}

	return savedSearches, nil
}

// Update updates an existing saved search, validating ownership
func (r *SavedSearchRepository) Update(ctx context.Context, userID int64, id int64, savedSearchEntity *savedsearch.SavedSearch) error {
	return r.Save(ctx, userID, savedSearchEntity)
}

// Delete deletes a saved search, validating ownership (soft delete)
func (r *SavedSearchRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingSavedSearch, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingSavedSearch == nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE saved_searches
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete saved search: %w", err)
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

// Exists checks if a saved search exists and belongs to the user
func (r *SavedSearchRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM saved_searches
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check saved search existence: %w", err)
	}

	return exists, nil
}

// Ensure SavedSearchRepository implements ISavedSearchRepository
var _ secondary.ISavedSearchRepository = (*SavedSearchRepository)(nil)

