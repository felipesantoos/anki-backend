package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// MediaRepository implements IMediaRepository using PostgreSQL
type MediaRepository struct {
	db *sql.DB
}

// NewMediaRepository creates a new MediaRepository instance
func NewMediaRepository(db *sql.DB) secondary.IMediaRepository {
	return &MediaRepository{
		db: db,
	}
}

// Save saves or updates a media in the database
func (r *MediaRepository) Save(ctx context.Context, userID int64, mediaEntity *media.Media) error {
	model := mappers.MediaToModel(mediaEntity)

	if mediaEntity.GetID() == 0 {
		// Insert new media
		query := `
			INSERT INTO media (user_id, filename, hash, size, mime_type, storage_path, created_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}

		var deletedAt sql.NullTime
		if model.DeletedAt.Valid {
			deletedAt = model.DeletedAt
		}

		var mediaID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Filename,
			model.Hash,
			model.Size,
			model.MimeType,
			model.StoragePath,
			model.CreatedAt,
			deletedAt,
		).Scan(&mediaID)
		if err != nil {
			return fmt.Errorf("failed to create media: %w", err)
		}

		mediaEntity.SetID(mediaID)
		return nil
	}

	// Update existing media - validate ownership first
	existingMedia, err := r.FindByID(ctx, userID, mediaEntity.GetID())
	if err != nil {
		return err
	}
	if existingMedia == nil {
		return ownership.ErrResourceNotFound
	}

	// Update media
	query := `
		UPDATE media
		SET filename = $1, hash = $2, size = $3, mime_type = $4, storage_path = $5, deleted_at = $6
		WHERE id = $7 AND user_id = $8 AND deleted_at IS NULL
	`

	var deletedAt sql.NullTime
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Filename,
		model.Hash,
		model.Size,
		model.MimeType,
		model.StoragePath,
		deletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update media: %w", err)
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

// FindByID finds a media by ID, filtering by userID to ensure ownership
func (r *MediaRepository) FindByID(ctx context.Context, userID int64, id int64) (*media.Media, error) {
	query := `
		SELECT id, user_id, filename, hash, size, mime_type, storage_path, created_at, deleted_at
		FROM media
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.MediaModel
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Filename,
		&model.Hash,
		&model.Size,
		&model.MimeType,
		&model.StoragePath,
		&model.CreatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find media: %w", err)
	}

	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.MediaToDomain(&model)
}

// FindByUserID finds all media for a user
func (r *MediaRepository) FindByUserID(ctx context.Context, userID int64) ([]*media.Media, error) {
	query := `
		SELECT id, user_id, filename, hash, size, mime_type, storage_path, created_at, deleted_at
		FROM media
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find media by user ID: %w", err)
	}
	defer rows.Close()

	var medias []*media.Media
	for rows.Next() {
		var model models.MediaModel
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Filename,
			&model.Hash,
			&model.Size,
			&model.MimeType,
			&model.StoragePath,
			&model.CreatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}

		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		mediaEntity, err := mappers.MediaToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert media to domain: %w", err)
		}
		medias = append(medias, mediaEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating media: %w", err)
	}

	return medias, nil
}

// Update updates an existing media, validating ownership
func (r *MediaRepository) Update(ctx context.Context, userID int64, id int64, mediaEntity *media.Media) error {
	return r.Save(ctx, userID, mediaEntity)
}

// Delete deletes a media, validating ownership (soft delete)
func (r *MediaRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingMedia, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingMedia == nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE media
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
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

// Exists checks if a media exists and belongs to the user
func (r *MediaRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM media
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check media existence: %w", err)
	}

	return exists, nil
}

// FindByHash finds a media by hash for a user
func (r *MediaRepository) FindByHash(ctx context.Context, userID int64, hash string) (*media.Media, error) {
	query := `
		SELECT id, user_id, filename, hash, size, mime_type, storage_path, created_at, deleted_at
		FROM media
		WHERE user_id = $1 AND hash = $2 AND deleted_at IS NULL
	`

	var model models.MediaModel
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, hash).Scan(
		&model.ID,
		&model.UserID,
		&model.Filename,
		&model.Hash,
		&model.Size,
		&model.MimeType,
		&model.StoragePath,
		&model.CreatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find media by hash: %w", err)
	}

	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	return mappers.MediaToDomain(&model)
}

// FindByFilename finds a media by filename for a user
func (r *MediaRepository) FindByFilename(ctx context.Context, userID int64, filename string) (*media.Media, error) {
	query := `
		SELECT id, user_id, filename, hash, size, mime_type, storage_path, created_at, deleted_at
		FROM media
		WHERE user_id = $1 AND filename = $2 AND deleted_at IS NULL
	`

	var model models.MediaModel
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, filename).Scan(
		&model.ID,
		&model.UserID,
		&model.Filename,
		&model.Hash,
		&model.Size,
		&model.MimeType,
		&model.StoragePath,
		&model.CreatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find media by filename: %w", err)
	}

	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	return mappers.MediaToDomain(&model)
}

// Ensure MediaRepository implements IMediaRepository
var _ secondary.IMediaRepository = (*MediaRepository)(nil)

