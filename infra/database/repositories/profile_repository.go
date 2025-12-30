package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// ProfileRepository implements IProfileRepository using PostgreSQL
type ProfileRepository struct {
	db *sql.DB
}

// NewProfileRepository creates a new ProfileRepository instance
func NewProfileRepository(db *sql.DB) secondary.IProfileRepository {
	return &ProfileRepository{
		db: db,
	}
}

// Save saves or updates a profile in the database
func (r *ProfileRepository) Save(ctx context.Context, userID int64, profileEntity *profile.Profile) error {
	model := mappers.ProfileToModel(profileEntity)

	if profileEntity.GetID() == 0 {
		// Insert new profile
		query := `
			INSERT INTO profiles (user_id, name, ankiweb_sync_enabled, ankiweb_username, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var ankiWebUsername sql.NullString
		if model.AnkiWebUsername.Valid {
			ankiWebUsername = model.AnkiWebUsername
		}

		var deletedAt sql.NullTime
		if model.DeletedAt.Valid {
			deletedAt = model.DeletedAt
		}

		var profileID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Name,
			model.AnkiWebSyncEnabled,
			ankiWebUsername,
			model.CreatedAt,
			model.UpdatedAt,
			deletedAt,
		).Scan(&profileID)
		if err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		profileEntity.SetID(profileID)
		return nil
	}

	// Update existing profile - validate ownership first
	existingProfile, err := r.FindByID(ctx, userID, profileEntity.GetID())
	if err != nil {
		return err
	}
	if existingProfile == nil {
		return ownership.ErrResourceNotFound
	}

	// Update profile
	query := `
		UPDATE profiles
		SET name = $1, ankiweb_sync_enabled = $2, ankiweb_username = $3, updated_at = $4, deleted_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`

	now := time.Now()
	model.UpdatedAt = now

	var ankiWebUsername sql.NullString
	if model.AnkiWebUsername.Valid {
		ankiWebUsername = model.AnkiWebUsername
	}

	var deletedAt sql.NullTime
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Name,
		model.AnkiWebSyncEnabled,
		ankiWebUsername,
		model.UpdatedAt,
		deletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
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

// FindByID finds a profile by ID, filtering by userID to ensure ownership
func (r *ProfileRepository) FindByID(ctx context.Context, userID int64, id int64) (*profile.Profile, error) {
	query := `
		SELECT id, user_id, name, ankiweb_sync_enabled, ankiweb_username, created_at, updated_at, deleted_at
		FROM profiles
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.ProfileModel
	var ankiWebUsername sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.AnkiWebSyncEnabled,
		&ankiWebUsername,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find profile: %w", err)
	}

	model.AnkiWebUsername = ankiWebUsername
	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.ProfileToDomain(&model)
}

// FindByUserID finds all profiles for a user
func (r *ProfileRepository) FindByUserID(ctx context.Context, userID int64) ([]*profile.Profile, error) {
	query := `
		SELECT id, user_id, name, ankiweb_sync_enabled, ankiweb_username, created_at, updated_at, deleted_at
		FROM profiles
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find profiles by user ID: %w", err)
	}
	defer rows.Close()

	var profiles []*profile.Profile
	for rows.Next() {
		var model models.ProfileModel
		var ankiWebUsername sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.AnkiWebSyncEnabled,
			&ankiWebUsername,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}

		model.AnkiWebUsername = ankiWebUsername
		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		profileEntity, err := mappers.ProfileToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert profile to domain: %w", err)
		}
		profiles = append(profiles, profileEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating profiles: %w", err)
	}

	return profiles, nil
}

// Update updates an existing profile, validating ownership
func (r *ProfileRepository) Update(ctx context.Context, userID int64, id int64, profileEntity *profile.Profile) error {
	return r.Save(ctx, userID, profileEntity)
}

// Delete deletes a profile, validating ownership (soft delete)
func (r *ProfileRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingProfile, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingProfile == nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE profiles
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
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

// Exists checks if a profile exists and belongs to the user
func (r *ProfileRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM profiles
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check profile existence: %w", err)
	}

	return exists, nil
}

// FindByName finds a profile by name, filtering by userID to ensure ownership
func (r *ProfileRepository) FindByName(ctx context.Context, userID int64, name string) (*profile.Profile, error) {
	query := `
		SELECT id, user_id, name, ankiweb_sync_enabled, ankiweb_username, created_at, updated_at, deleted_at
		FROM profiles
		WHERE name = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.ProfileModel
	err := r.db.QueryRowContext(ctx, query, name, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.AnkiWebSyncEnabled,
		&model.AnkiWebUsername,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find profile by name: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.ProfileToDomain(&model)
}

// Ensure ProfileRepository implements IProfileRepository
var _ secondary.IProfileRepository = (*ProfileRepository)(nil)

