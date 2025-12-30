package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	addon "github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// AddOnRepository implements IAddOnRepository using PostgreSQL
type AddOnRepository struct {
	db *sql.DB
}

// NewAddOnRepository creates a new AddOnRepository instance
func NewAddOnRepository(db *sql.DB) secondary.IAddOnRepository {
	return &AddOnRepository{
		db: db,
	}
}

// Save saves or updates an add-on in the database
func (r *AddOnRepository) Save(ctx context.Context, userID int64, addOnEntity *addon.AddOn) error {
	model := mappers.AddOnToModel(addOnEntity)

	if addOnEntity.GetID() == 0 {
		// Insert new add-on
		query := `
			INSERT INTO add_ons (user_id, code, name, version, enabled, config_json, installed_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`

		now := time.Now()
		if model.InstalledAt.IsZero() {
			model.InstalledAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var addOnID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Code,
			model.Name,
			model.Version,
			model.Enabled,
			model.ConfigJSON,
			model.InstalledAt,
			model.UpdatedAt,
		).Scan(&addOnID)
		if err != nil {
			return fmt.Errorf("failed to create add-on: %w", err)
		}

		addOnEntity.SetID(addOnID)
		return nil
	}

	// Update existing add-on - validate ownership first
	existingAddOn, err := r.FindByID(ctx, userID, addOnEntity.GetID())
	if err != nil {
		return err
	}
	if existingAddOn == nil {
		return ownership.ErrResourceNotFound
	}

	// Update add-on
	query := `
		UPDATE add_ons
		SET code = $1, name = $2, version = $3, enabled = $4, config_json = $5, updated_at = $6
		WHERE id = $7 AND user_id = $8
	`

	now := time.Now()
	model.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		model.Code,
		model.Name,
		model.Version,
		model.Enabled,
		model.ConfigJSON,
		model.UpdatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update add-on: %w", err)
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

// FindByID finds an add-on by ID, filtering by userID to ensure ownership
func (r *AddOnRepository) FindByID(ctx context.Context, userID int64, id int64) (*addon.AddOn, error) {
	query := `
		SELECT id, user_id, code, name, version, enabled, config_json, installed_at, updated_at
		FROM add_ons
		WHERE id = $1 AND user_id = $2
	`

	var model models.AddOnModel
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Code,
		&model.Name,
		&model.Version,
		&model.Enabled,
		&model.ConfigJSON,
		&model.InstalledAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find add-on: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.AddOnToDomain(&model)
}

// FindByUserID finds all add-ons for a user
func (r *AddOnRepository) FindByUserID(ctx context.Context, userID int64) ([]*addon.AddOn, error) {
	query := `
		SELECT id, user_id, code, name, version, enabled, config_json, installed_at, updated_at
		FROM add_ons
		WHERE user_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find add-ons by user ID: %w", err)
	}
	defer rows.Close()

	var addOns []*addon.AddOn
	for rows.Next() {
		var model models.AddOnModel

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Code,
			&model.Name,
			&model.Version,
			&model.Enabled,
			&model.ConfigJSON,
			&model.InstalledAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan add-on: %w", err)
		}

		addOnEntity, err := mappers.AddOnToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert add-on to domain: %w", err)
		}
		addOns = append(addOns, addOnEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating add-ons: %w", err)
	}

	return addOns, nil
}

// Update updates an existing add-on, validating ownership
func (r *AddOnRepository) Update(ctx context.Context, userID int64, id int64, addOnEntity *addon.AddOn) error {
	return r.Save(ctx, userID, addOnEntity)
}

// Delete deletes an add-on, validating ownership
func (r *AddOnRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingAddOn, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingAddOn == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (add_ons don't have soft delete)
	query := `DELETE FROM add_ons WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete add-on: %w", err)
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

// Exists checks if an add-on exists and belongs to the user
func (r *AddOnRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM add_ons
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check add-on existence: %w", err)
	}

	return exists, nil
}

// FindByCode finds an add-on by code, filtering by userID to ensure ownership
func (r *AddOnRepository) FindByCode(ctx context.Context, userID int64, code string) (*addon.AddOn, error) {
	query := `
		SELECT id, user_id, code, name, version, enabled, config_json, installed_at, updated_at
		FROM add_ons
		WHERE code = $1 AND user_id = $2
	`

	var model models.AddOnModel
	err := r.db.QueryRowContext(ctx, query, code, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Code,
		&model.Name,
		&model.Version,
		&model.Enabled,
		&model.ConfigJSON,
		&model.InstalledAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find add-on by code: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.AddOnToDomain(&model)
}

// FindEnabled finds all enabled add-ons for a user
func (r *AddOnRepository) FindEnabled(ctx context.Context, userID int64) ([]*addon.AddOn, error) {
	query := `
		SELECT id, user_id, code, name, version, enabled, config_json, installed_at, updated_at
		FROM add_ons
		WHERE user_id = $1 AND enabled = true
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find enabled add-ons: %w", err)
	}
	defer rows.Close()

	var addOns []*addon.AddOn
	for rows.Next() {
		var model models.AddOnModel
		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Code,
			&model.Name,
			&model.Version,
			&model.Enabled,
			&model.ConfigJSON,
			&model.InstalledAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan add-on: %w", err)
		}

		addOnEntity, err := mappers.AddOnToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert add-on to domain: %w", err)
		}
		addOns = append(addOns, addOnEntity)
	}

	return addOns, nil
}

// Ensure AddOnRepository implements IAddOnRepository
var _ secondary.IAddOnRepository = (*AddOnRepository)(nil)

