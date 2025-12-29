package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// SyncMetaRepository implements ISyncMetaRepository using PostgreSQL
type SyncMetaRepository struct {
	db *sql.DB
}

// NewSyncMetaRepository creates a new SyncMetaRepository instance
func NewSyncMetaRepository(db *sql.DB) secondary.ISyncMetaRepository {
	return &SyncMetaRepository{
		db: db,
	}
}

// Save saves or updates sync metadata in the database
func (r *SyncMetaRepository) Save(ctx context.Context, userID int64, syncMetaEntity *syncmeta.SyncMeta) error {
	model := mappers.SyncMetaToModel(syncMetaEntity)

	if syncMetaEntity.GetID() == 0 {
		// Insert new sync meta
		query := `
			INSERT INTO sync_meta (user_id, client_id, last_sync, last_sync_usn, created_at, updated_at)
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

		var syncMetaID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.ClientID,
			model.LastSync,
			model.LastSyncUSN,
			model.CreatedAt,
			model.UpdatedAt,
		).Scan(&syncMetaID)
		if err != nil {
			return fmt.Errorf("failed to create sync meta: %w", err)
		}

		syncMetaEntity.SetID(syncMetaID)
		return nil
	}

	// Update existing sync meta - validate ownership first
	existingSyncMeta, err := r.FindByID(ctx, userID, syncMetaEntity.GetID())
	if err != nil {
		return err
	}
	if existingSyncMeta == nil {
		return ownership.ErrResourceNotFound
	}

	// Update sync meta
	query := `
		UPDATE sync_meta
		SET client_id = $1, last_sync = $2, last_sync_usn = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`

	now := time.Now()
	model.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		model.ClientID,
		model.LastSync,
		model.LastSyncUSN,
		model.UpdatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update sync meta: %w", err)
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

// FindByID finds sync metadata by ID, filtering by userID to ensure ownership
func (r *SyncMetaRepository) FindByID(ctx context.Context, userID int64, id int64) (*syncmeta.SyncMeta, error) {
	query := `
		SELECT id, user_id, client_id, last_sync, last_sync_usn, created_at, updated_at
		FROM sync_meta
		WHERE id = $1 AND user_id = $2
	`

	var model models.SyncMetaModel
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.ClientID,
		&model.LastSync,
		&model.LastSyncUSN,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find sync meta: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.SyncMetaToDomain(&model)
}

// FindByUserID finds all sync metadata for a user
func (r *SyncMetaRepository) FindByUserID(ctx context.Context, userID int64) ([]*syncmeta.SyncMeta, error) {
	query := `
		SELECT id, user_id, client_id, last_sync, last_sync_usn, created_at, updated_at
		FROM sync_meta
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find sync meta by user ID: %w", err)
	}
	defer rows.Close()

	var syncMetas []*syncmeta.SyncMeta
	for rows.Next() {
		var model models.SyncMetaModel

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.ClientID,
			&model.LastSync,
			&model.LastSyncUSN,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sync meta: %w", err)
		}

		syncMetaEntity, err := mappers.SyncMetaToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert sync meta to domain: %w", err)
		}
		syncMetas = append(syncMetas, syncMetaEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sync meta: %w", err)
	}

	return syncMetas, nil
}

// Update updates existing sync metadata, validating ownership
func (r *SyncMetaRepository) Update(ctx context.Context, userID int64, id int64, syncMetaEntity *syncmeta.SyncMeta) error {
	return r.Save(ctx, userID, syncMetaEntity)
}

// Delete deletes sync metadata, validating ownership
func (r *SyncMetaRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingSyncMeta, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingSyncMeta == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (sync_meta doesn't have soft delete)
	query := `DELETE FROM sync_meta WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sync meta: %w", err)
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

// Exists checks if sync metadata exists and belongs to the user
func (r *SyncMetaRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM sync_meta
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check sync meta existence: %w", err)
	}

	return exists, nil
}

// FindByClientID finds sync metadata by client ID for a user
func (r *SyncMetaRepository) FindByClientID(ctx context.Context, userID int64, clientID string) (*syncmeta.SyncMeta, error) {
	query := `
		SELECT id, user_id, client_id, last_sync, last_sync_usn, created_at, updated_at
		FROM sync_meta
		WHERE user_id = $1 AND client_id = $2
	`

	var model models.SyncMetaModel
	err := r.db.QueryRowContext(ctx, query, userID, clientID).Scan(
		&model.ID,
		&model.UserID,
		&model.ClientID,
		&model.LastSync,
		&model.LastSyncUSN,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find sync meta by client ID: %w", err)
	}

	return mappers.SyncMetaToDomain(&model)
}

// Ensure SyncMetaRepository implements ISyncMetaRepository
var _ secondary.ISyncMetaRepository = (*SyncMetaRepository)(nil)

