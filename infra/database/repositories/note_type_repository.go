package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// NoteTypeRepository implements INoteTypeRepository using PostgreSQL
type NoteTypeRepository struct {
	db *sql.DB
}

// NewNoteTypeRepository creates a new NoteTypeRepository instance
func NewNoteTypeRepository(db *sql.DB) secondary.INoteTypeRepository {
	return &NoteTypeRepository{
		db: db,
	}
}

// Save saves or updates a note type in the database
func (r *NoteTypeRepository) Save(ctx context.Context, userID int64, noteTypeEntity *notetype.NoteType) error {
	model := mappers.NoteTypeToModel(noteTypeEntity)

	if noteTypeEntity.GetID() == 0 {
		// Insert new note type
		query := `
			INSERT INTO note_types (user_id, name, fields_json, card_types_json, templates_json, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

		var noteTypeID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Name,
			model.FieldsJSON,
			model.CardTypesJSON,
			model.TemplatesJSON,
			model.CreatedAt,
			model.UpdatedAt,
			deletedAt,
		).Scan(&noteTypeID)
		if err != nil {
			return fmt.Errorf("failed to create note type: %w", err)
		}

		noteTypeEntity.SetID(noteTypeID)
		return nil
	}

	// Update existing note type - validate ownership first
	existingNoteType, err := r.FindByID(ctx, userID, noteTypeEntity.GetID())
	if err != nil {
		return err
	}
	if existingNoteType == nil {
		return ownership.ErrResourceNotFound
	}

	// Update note type
	query := `
		UPDATE note_types
		SET name = $1, fields_json = $2, card_types_json = $3, templates_json = $4, updated_at = $5, deleted_at = $6
		WHERE id = $7 AND user_id = $8 AND deleted_at IS NULL
	`

	now := time.Now()
	model.UpdatedAt = now

	var deletedAt sql.NullTime
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Name,
		model.FieldsJSON,
		model.CardTypesJSON,
		model.TemplatesJSON,
		model.UpdatedAt,
		deletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update note type: %w", err)
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

// FindByID finds a note type by ID, filtering by userID to ensure ownership
func (r *NoteTypeRepository) FindByID(ctx context.Context, userID int64, id int64) (*notetype.NoteType, error) {
	query := `
		SELECT id, user_id, name, fields_json, card_types_json, templates_json, created_at, updated_at, deleted_at
		FROM note_types
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.NoteTypeModel
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.FieldsJSON,
		&model.CardTypesJSON,
		&model.TemplatesJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find note type: %w", err)
	}

	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.NoteTypeToDomain(&model)
}

// FindByUserID finds all note types for a user, with optional search filter
func (r *NoteTypeRepository) FindByUserID(ctx context.Context, userID int64, search string) ([]*notetype.NoteType, error) {
	var query string
	var args []interface{}

	if search != "" {
		// Escape special characters for ILIKE
		escapedSearch := strings.ReplaceAll(search, "%", "\\%")
		escapedSearch = strings.ReplaceAll(escapedSearch, "_", "\\_")
		query = `
			SELECT id, user_id, name, fields_json, card_types_json, templates_json, created_at, updated_at, deleted_at
			FROM note_types
			WHERE user_id = $1 AND deleted_at IS NULL AND name ILIKE $2
			ORDER BY name ASC
		`
		args = []interface{}{userID, "%" + escapedSearch + "%"}
	} else {
		query = `
			SELECT id, user_id, name, fields_json, card_types_json, templates_json, created_at, updated_at, deleted_at
			FROM note_types
			WHERE user_id = $1 AND deleted_at IS NULL
			ORDER BY name ASC
		`
		args = []interface{}{userID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find note types by user ID: %w", err)
	}
	defer rows.Close()

	var noteTypes []*notetype.NoteType
	for rows.Next() {
		var model models.NoteTypeModel
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.FieldsJSON,
			&model.CardTypesJSON,
			&model.TemplatesJSON,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note type: %w", err)
		}

		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		noteTypeEntity, err := mappers.NoteTypeToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert note type to domain: %w", err)
		}
		noteTypes = append(noteTypes, noteTypeEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating note types: %w", err)
	}

	return noteTypes, nil
}

// Update updates an existing note type, validating ownership
func (r *NoteTypeRepository) Update(ctx context.Context, userID int64, id int64, noteTypeEntity *notetype.NoteType) error {
	return r.Save(ctx, userID, noteTypeEntity)
}

// Delete deletes a note type, validating ownership (soft delete)
func (r *NoteTypeRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingNoteType, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingNoteType == nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE note_types
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete note type: %w", err)
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

// Exists checks if a note type exists and belongs to the user
func (r *NoteTypeRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM note_types
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check note type existence: %w", err)
	}

	return exists, nil
}

// FindByName finds a note type by name for a user
func (r *NoteTypeRepository) FindByName(ctx context.Context, userID int64, name string) (*notetype.NoteType, error) {
	query := `
		SELECT id, user_id, name, fields_json, card_types_json, templates_json, created_at, updated_at, deleted_at
		FROM note_types
		WHERE user_id = $1 AND name = $2 AND deleted_at IS NULL
	`

	var model models.NoteTypeModel
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.FieldsJSON,
		&model.CardTypesJSON,
		&model.TemplatesJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find note type by name: %w", err)
	}

	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.NoteTypeToDomain(&model)
}

// ExistsByName checks if a note type with the given name exists for the user
func (r *NoteTypeRepository) ExistsByName(ctx context.Context, userID int64, name string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM note_types
			WHERE user_id = $1 AND name = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check note type existence by name: %w", err)
	}

	return exists, nil
}

// Ensure NoteTypeRepository implements INoteTypeRepository
var _ secondary.INoteTypeRepository = (*NoteTypeRepository)(nil)

