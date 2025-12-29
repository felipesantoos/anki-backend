package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	filtereddeck "github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// FilteredDeckRepository implements IFilteredDeckRepository using PostgreSQL
type FilteredDeckRepository struct {
	db *sql.DB
}

// NewFilteredDeckRepository creates a new FilteredDeckRepository instance
func NewFilteredDeckRepository(db *sql.DB) secondary.IFilteredDeckRepository {
	return &FilteredDeckRepository{
		db: db,
	}
}

// Save saves or updates a filtered deck in the database
func (r *FilteredDeckRepository) Save(ctx context.Context, userID int64, filteredDeckEntity *filtereddeck.FilteredDeck) error {
	model := mappers.FilteredDeckToModel(filteredDeckEntity)

	if filteredDeckEntity.GetID() == 0 {
		// Insert new filtered deck
		query := `
			INSERT INTO filtered_decks (user_id, name, search_filter, second_filter, limit_cards, order_by, reschedule, created_at, updated_at, last_rebuild_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var secondFilter interface{}
		if model.SecondFilter.Valid {
			secondFilter = model.SecondFilter.String
		}

		var lastRebuildAt interface{}
		if model.LastRebuildAt.Valid {
			lastRebuildAt = model.LastRebuildAt.Time
		}

		var deletedAt interface{}
		if model.DeletedAt.Valid {
			deletedAt = model.DeletedAt.Time
		}

		var filteredDeckID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Name,
			model.SearchFilter,
			secondFilter,
			model.LimitCards,
			model.OrderBy,
			model.Reschedule,
			model.CreatedAt,
			model.UpdatedAt,
			lastRebuildAt,
			deletedAt,
		).Scan(&filteredDeckID)
		if err != nil {
			return fmt.Errorf("failed to create filtered deck: %w", err)
		}

		filteredDeckEntity.SetID(filteredDeckID)
		return nil
	}

	// Update existing filtered deck - validate ownership first
	existingFilteredDeck, err := r.FindByID(ctx, userID, filteredDeckEntity.GetID())
	if err != nil {
		return err
	}
	if existingFilteredDeck == nil {
		return ownership.ErrResourceNotFound
	}

	// Update filtered deck
	query := `
		UPDATE filtered_decks
		SET name = $1, search_filter = $2, second_filter = $3, limit_cards = $4, order_by = $5, reschedule = $6, updated_at = $7, last_rebuild_at = $8, deleted_at = $9
		WHERE id = $10 AND user_id = $11 AND deleted_at IS NULL
	`

	now := time.Now()
	model.UpdatedAt = now

	var secondFilter interface{}
	if model.SecondFilter.Valid {
		secondFilter = model.SecondFilter.String
	}

	var lastRebuildAt interface{}
	if model.LastRebuildAt.Valid {
		lastRebuildAt = model.LastRebuildAt.Time
	}

	var deletedAt interface{}
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt.Time
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Name,
		model.SearchFilter,
		secondFilter,
		model.LimitCards,
		model.OrderBy,
		model.Reschedule,
		model.UpdatedAt,
		lastRebuildAt,
		deletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update filtered deck: %w", err)
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

// FindByID finds a filtered deck by ID, filtering by userID to ensure ownership
func (r *FilteredDeckRepository) FindByID(ctx context.Context, userID int64, id int64) (*filtereddeck.FilteredDeck, error) {
	query := `
		SELECT id, user_id, name, search_filter, second_filter, limit_cards, order_by, reschedule, created_at, updated_at, last_rebuild_at, deleted_at
		FROM filtered_decks
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.FilteredDeckModel
	var secondFilter sql.NullString
	var lastRebuildAt sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.SearchFilter,
		&secondFilter,
		&model.LimitCards,
		&model.OrderBy,
		&model.Reschedule,
		&model.CreatedAt,
		&model.UpdatedAt,
		&lastRebuildAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find filtered deck: %w", err)
	}

	model.SecondFilter = secondFilter
	model.LastRebuildAt = lastRebuildAt
	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.FilteredDeckToDomain(&model)
}

// FindByUserID finds all filtered decks for a user
func (r *FilteredDeckRepository) FindByUserID(ctx context.Context, userID int64) ([]*filtereddeck.FilteredDeck, error) {
	query := `
		SELECT id, user_id, name, search_filter, second_filter, limit_cards, order_by, reschedule, created_at, updated_at, last_rebuild_at, deleted_at
		FROM filtered_decks
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find filtered decks by user ID: %w", err)
	}
	defer rows.Close()

	var filteredDecks []*filtereddeck.FilteredDeck
	for rows.Next() {
		var model models.FilteredDeckModel
		var secondFilter sql.NullString
		var lastRebuildAt sql.NullTime
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.SearchFilter,
			&secondFilter,
			&model.LimitCards,
			&model.OrderBy,
			&model.Reschedule,
			&model.CreatedAt,
			&model.UpdatedAt,
			&lastRebuildAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan filtered deck: %w", err)
		}

		model.SecondFilter = secondFilter
		model.LastRebuildAt = lastRebuildAt
		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		filteredDeckEntity, err := mappers.FilteredDeckToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filtered deck to domain: %w", err)
		}
		filteredDecks = append(filteredDecks, filteredDeckEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating filtered decks: %w", err)
	}

	return filteredDecks, nil
}

// Update updates an existing filtered deck, validating ownership
func (r *FilteredDeckRepository) Update(ctx context.Context, userID int64, id int64, filteredDeckEntity *filtereddeck.FilteredDeck) error {
	return r.Save(ctx, userID, filteredDeckEntity)
}

// Delete deletes a filtered deck, validating ownership (soft delete)
func (r *FilteredDeckRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingFilteredDeck, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingFilteredDeck == nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE filtered_decks
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete filtered deck: %w", err)
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

// Exists checks if a filtered deck exists and belongs to the user
func (r *FilteredDeckRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM filtered_decks
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check filtered deck existence: %w", err)
	}

	return exists, nil
}

// Ensure FilteredDeckRepository implements IFilteredDeckRepository
var _ secondary.IFilteredDeckRepository = (*FilteredDeckRepository)(nil)

