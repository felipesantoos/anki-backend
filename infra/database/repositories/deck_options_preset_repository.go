package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	deckoptionspreset "github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// DeckOptionsPresetRepository implements IDeckOptionsPresetRepository using PostgreSQL
type DeckOptionsPresetRepository struct {
	db *sql.DB
}

// NewDeckOptionsPresetRepository creates a new DeckOptionsPresetRepository instance
func NewDeckOptionsPresetRepository(db *sql.DB) secondary.IDeckOptionsPresetRepository {
	return &DeckOptionsPresetRepository{
		db: db,
	}
}

// Save saves or updates a deck options preset in the database
func (r *DeckOptionsPresetRepository) Save(ctx context.Context, userID int64, presetEntity *deckoptionspreset.DeckOptionsPreset) error {
	model := mappers.DeckOptionsPresetToModel(presetEntity)

	if presetEntity.GetID() == 0 {
		// Insert new preset
		query := `
			INSERT INTO deck_options_presets (user_id, name, options_json, created_at, updated_at, deleted_at)
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

		var presetID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Name,
			model.OptionsJSON,
			model.CreatedAt,
			model.UpdatedAt,
			deletedAt,
		).Scan(&presetID)
		if err != nil {
			return fmt.Errorf("failed to create deck options preset: %w", err)
		}

		presetEntity.SetID(presetID)
		return nil
	}

	// Update existing preset - validate ownership first
	existingPreset, err := r.FindByID(ctx, userID, presetEntity.GetID())
	if err != nil {
		return err
	}
	if existingPreset == nil {
		return ownership.ErrResourceNotFound
	}

	// Update preset
	query := `
		UPDATE deck_options_presets
		SET name = $1, options_json = $2, updated_at = $3, deleted_at = $4
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
		model.OptionsJSON,
		model.UpdatedAt,
		deletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update deck options preset: %w", err)
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

// FindByID finds a deck options preset by ID, filtering by userID to ensure ownership
func (r *DeckOptionsPresetRepository) FindByID(ctx context.Context, userID int64, id int64) (*deckoptionspreset.DeckOptionsPreset, error) {
	query := `
		SELECT id, user_id, name, options_json, created_at, updated_at, deleted_at
		FROM deck_options_presets
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.DeckOptionsPresetModel
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.OptionsJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find deck options preset: %w", err)
	}

	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.DeckOptionsPresetToDomain(&model)
}

// FindByUserID finds all deck options presets for a user
func (r *DeckOptionsPresetRepository) FindByUserID(ctx context.Context, userID int64) ([]*deckoptionspreset.DeckOptionsPreset, error) {
	query := `
		SELECT id, user_id, name, options_json, created_at, updated_at, deleted_at
		FROM deck_options_presets
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find deck options presets by user ID: %w", err)
	}
	defer rows.Close()

	var presets []*deckoptionspreset.DeckOptionsPreset
	for rows.Next() {
		var model models.DeckOptionsPresetModel
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.OptionsJSON,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck options preset: %w", err)
		}

		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		presetEntity, err := mappers.DeckOptionsPresetToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert deck options preset to domain: %w", err)
		}
		presets = append(presets, presetEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck options presets: %w", err)
	}

	return presets, nil
}

// Update updates an existing deck options preset, validating ownership
func (r *DeckOptionsPresetRepository) Update(ctx context.Context, userID int64, id int64, presetEntity *deckoptionspreset.DeckOptionsPreset) error {
	return r.Save(ctx, userID, presetEntity)
}

// Delete deletes a deck options preset, validating ownership (soft delete)
func (r *DeckOptionsPresetRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingPreset, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingPreset == nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE deck_options_presets
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete deck options preset: %w", err)
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

// Exists checks if a deck options preset exists and belongs to the user
func (r *DeckOptionsPresetRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM deck_options_presets
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check deck options preset existence: %w", err)
	}

	return exists, nil
}

// FindByName finds a deck options preset by name, filtering by userID to ensure ownership
func (r *DeckOptionsPresetRepository) FindByName(ctx context.Context, userID int64, name string) (*deckoptionspreset.DeckOptionsPreset, error) {
	query := `
		SELECT id, user_id, name, options_json, created_at, updated_at, deleted_at
		FROM deck_options_presets
		WHERE name = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.DeckOptionsPresetModel
	err := r.db.QueryRowContext(ctx, query, name, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.OptionsJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find deck options preset by name: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.DeckOptionsPresetToDomain(&model)
}

// Ensure DeckOptionsPresetRepository implements IDeckOptionsPresetRepository
var _ secondary.IDeckOptionsPresetRepository = (*DeckOptionsPresetRepository)(nil)

