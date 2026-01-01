package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/services/search"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// NoteRepository implements INoteRepository using PostgreSQL
type NoteRepository struct {
	db *sql.DB
}

// NewNoteRepository creates a new NoteRepository instance
func NewNoteRepository(db *sql.DB) secondary.INoteRepository {
	return &NoteRepository{
		db: db,
	}
}

// Save saves or updates a note in the database
func (r *NoteRepository) Save(ctx context.Context, userID int64, noteEntity *note.Note) error {
	model := mappers.NoteToModel(noteEntity)

	if noteEntity.GetID() == 0 {
		// Insert new note
		query := `
			INSERT INTO notes (user_id, guid, note_type_id, fields_json, tags, marked, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5::TEXT[], $6, $7, $8, $9)
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

		var noteID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.GUID,
			model.NoteTypeID,
			model.FieldsJSON,
			pq.Array(noteEntity.GetTags()), // Use pq.Array for PostgreSQL arrays directly from entity
			model.Marked,
			model.CreatedAt,
			model.UpdatedAt,
			deletedAt,
		).Scan(&noteID)
		if err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		noteEntity.SetID(noteID)
		return nil
	}

	// Update existing note - validate ownership first
	existingNote, err := r.FindByID(ctx, userID, noteEntity.GetID())
	if err != nil {
		return err
	}
	if existingNote == nil {
		return ownership.ErrResourceNotFound
	}

	// Update note
	query := `
		UPDATE notes
		SET guid = $1, note_type_id = $2, fields_json = $3, tags = $4::TEXT[], marked = $5, updated_at = $6, deleted_at = $7
		WHERE id = $8 AND user_id = $9 AND deleted_at IS NULL
	`

	now := time.Now()
	model.UpdatedAt = now

	var deletedAt sql.NullTime
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		model.GUID,
		model.NoteTypeID,
		model.FieldsJSON,
		pq.Array(noteEntity.GetTags()),
		model.Marked,
		model.UpdatedAt,
		deletedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
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

// FindByID finds a note by ID, filtering by userID to ensure ownership
func (r *NoteRepository) FindByID(ctx context.Context, userID int64, id int64) (*note.Note, error) {
	query := `
		SELECT id, user_id, guid, note_type_id, fields_json, tags, marked, created_at, updated_at, deleted_at
		FROM notes
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.NoteModel
	var tagsStr string
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.GUID,
		&model.NoteTypeID,
		&model.FieldsJSON,
		&tagsStr,
		&model.Marked,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find note: %w", err)
	}

	if tagsStr != "" {
		model.Tags = sql.NullString{String: tagsStr, Valid: true}
	}
	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.NoteToDomain(&model)
}

// FindByUserID finds all notes for a user with pagination
func (r *NoteRepository) FindByUserID(ctx context.Context, userID int64, limit int, offset int) ([]*note.Note, error) {
	query := `
		SELECT id, user_id, guid, note_type_id, fields_json, tags, marked, created_at, updated_at, deleted_at
		FROM notes
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find notes by user ID: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// scanNotes is a helper to scan multiple notes from rows
func (r *NoteRepository) scanNotes(rows *sql.Rows) ([]*note.Note, error) {
	var notes []*note.Note
	for rows.Next() {
		var model models.NoteModel
		var tagsStr string
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.GUID,
			&model.NoteTypeID,
			&model.FieldsJSON,
			&tagsStr,
			&model.Marked,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if tagsStr != "" {
			model.Tags = sql.NullString{String: tagsStr, Valid: true}
		}
		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		noteEntity, err := mappers.NoteToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert note to domain: %w", err)
		}
		notes = append(notes, noteEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notes: %w", err)
	}

	return notes, nil
}

// Update updates an existing note, validating ownership
func (r *NoteRepository) Update(ctx context.Context, userID int64, id int64, noteEntity *note.Note) error {
	return r.Save(ctx, userID, noteEntity)
}

// Delete deletes a note, validating ownership (soft delete)
func (r *NoteRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingNote, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingNote == nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE notes
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
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

// Exists checks if a note exists and belongs to the user
func (r *NoteRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM notes
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check note existence: %w", err)
	}

	return exists, nil
}

// FindByNoteTypeID finds all notes of a specific note type for a user with pagination
func (r *NoteRepository) FindByNoteTypeID(ctx context.Context, userID int64, noteTypeID int64, limit int, offset int) ([]*note.Note, error) {
	query := `
		SELECT id, user_id, guid, note_type_id, fields_json, tags, marked, created_at, updated_at, deleted_at
		FROM notes
		WHERE user_id = $1 AND note_type_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, noteTypeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find notes by note type ID: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// FindByDeckID finds all notes that have cards in a specific deck for a user with pagination
func (r *NoteRepository) FindByDeckID(ctx context.Context, userID int64, deckID int64, limit int, offset int) ([]*note.Note, error) {
	query := `
		SELECT DISTINCT n.id, n.user_id, n.guid, n.note_type_id, n.fields_json, n.tags, n.marked, n.created_at, n.updated_at, n.deleted_at
		FROM notes n
		JOIN cards c ON c.note_id = n.id
		WHERE n.user_id = $1 AND c.deck_id = $2 AND n.deleted_at IS NULL
		ORDER BY n.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, deckID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find notes by deck ID: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// FindByGUID finds a note by GUID, filtering by userID to ensure ownership
func (r *NoteRepository) FindByGUID(ctx context.Context, userID int64, guid string) (*note.Note, error) {
	query := `
		SELECT id, user_id, guid, note_type_id, fields_json, tags, marked, created_at, updated_at, deleted_at
		FROM notes
		WHERE guid = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.NoteModel
	var tagsStr string
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, guid, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.GUID,
		&model.NoteTypeID,
		&model.FieldsJSON,
		&tagsStr,
		&model.Marked,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find note by GUID: %w", err)
	}

	if tagsStr != "" {
		model.Tags = sql.NullString{String: tagsStr, Valid: true}
	}
	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.NoteToDomain(&model)
}

// FindByTags finds all notes containing any of the specified tags for a user with pagination
func (r *NoteRepository) FindByTags(ctx context.Context, userID int64, tags []string, limit int, offset int) ([]*note.Note, error) {
	if len(tags) == 0 {
		return []*note.Note{}, nil
	}

	query := `
		SELECT DISTINCT id, user_id, guid, note_type_id, fields_json, tags, marked, created_at, updated_at, deleted_at
		FROM notes
		WHERE user_id = $1 AND deleted_at IS NULL AND tags && $2::TEXT[]
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, pq.Array(tags), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find notes by tags: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// FindBySearch finds all notes containing the search text within fields_json for a user with pagination
// Searches case-insensitively within JSON field values using jsonb_each_text
func (r *NoteRepository) FindBySearch(ctx context.Context, userID int64, searchText string, limit int, offset int) ([]*note.Note, error) {
	if searchText == "" {
		return []*note.Note{}, nil
	}

	query := `
		SELECT DISTINCT id, user_id, guid, note_type_id, fields_json, tags, marked, created_at, updated_at, deleted_at
		FROM notes
		WHERE user_id = $1 
		  AND deleted_at IS NULL
		  AND EXISTS (
		      SELECT 1 FROM jsonb_each_text(fields_json) 
		      WHERE value ILIKE '%' || $2 || '%'
		  )
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, searchText, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find notes by search: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// FindByAdvancedSearch finds notes matching advanced search criteria
// Combines multiple filters: deck, tags, fields, card states, properties
func (r *NoteRepository) FindByAdvancedSearch(ctx context.Context, userID int64, query *search.SearchQuery, limit int, offset int) ([]*note.Note, error) {
	if query == nil {
		return []*note.Note{}, nil
	}

	// Build dynamic SQL query
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base conditions
	conditions = append(conditions, fmt.Sprintf("n.user_id = $%d", argIndex))
	args = append(args, userID)
	argIndex++

	conditions = append(conditions, "n.deleted_at IS NULL")

	// Deck filters
	if len(query.DecksInclude) > 0 {
		// Multiple decks with OR logic
		deckConditions := make([]string, len(query.DecksInclude))
		for i, deckName := range query.DecksInclude {
			deckConditions[i] = fmt.Sprintf("EXISTS (SELECT 1 FROM cards c JOIN decks d ON c.deck_id = d.id WHERE c.note_id = n.id AND d.user_id = $%d AND d.name = $%d)", argIndex, argIndex+1)
			args = append(args, userID, deckName)
			argIndex += 2
		}
		conditions = append(conditions, "("+strings.Join(deckConditions, " OR ")+")")
	}
	if len(query.DecksExclude) > 0 {
		for _, deckName := range query.DecksExclude {
			conditions = append(conditions, fmt.Sprintf("NOT EXISTS (SELECT 1 FROM cards c JOIN decks d ON c.deck_id = d.id WHERE c.note_id = n.id AND d.user_id = $%d AND d.name = $%d)", argIndex, argIndex+1))
			args = append(args, userID, deckName)
			argIndex += 2
		}
	}

	// Tag filters
	if len(query.TagsInclude) > 0 {
		conditions = append(conditions, fmt.Sprintf("n.tags && $%d::TEXT[]", argIndex))
		args = append(args, pq.Array(query.TagsInclude))
		argIndex++
	}
	if len(query.TagsExclude) > 0 {
		for _, tag := range query.TagsExclude {
			conditions = append(conditions, fmt.Sprintf("NOT ($%d::TEXT = ANY(n.tags))", argIndex))
			args = append(args, tag)
			argIndex++
		}
	}

	// Field searches (front:, back:, field:name:)
	for fieldName, searchText := range query.FieldSearches {
		// Escape search text for ILIKE
		escapedText := strings.ReplaceAll(searchText, "%", "\\%")
		escapedText = strings.ReplaceAll(escapedText, "_", "\\_")
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_each_text(n.fields_json) WHERE key = $%d AND value ILIKE $%d)", argIndex, argIndex+1))
		args = append(args, fieldName, "%"+escapedText+"%")
		argIndex += 2
	}

	// Text searches (general text in all fields)
	for _, textSearch := range query.TextSearches {
		if textSearch.IsNegated {
			continue // Handle negation separately if needed
		}

		if textSearch.IsExact {
			// Exact phrase search
			escapedText := strings.ReplaceAll(textSearch.Text, "%", "\\%")
			escapedText = strings.ReplaceAll(escapedText, "_", "\\_")
			conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_each_text(n.fields_json) WHERE value ILIKE $%d)", argIndex))
			args = append(args, "%"+escapedText+"%")
			argIndex++
		} else if textSearch.IsWildcard {
			// Wildcard search - convert * to % and _ to _
			wildcardText := strings.ReplaceAll(textSearch.Text, "*", "%")
			conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_each_text(n.fields_json) WHERE value ILIKE $%d)", argIndex))
			args = append(args, wildcardText)
			argIndex++
		} else if textSearch.IsRegex {
			// Regex search - use PostgreSQL ~ operator
			conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_each_text(n.fields_json) WHERE value ~ $%d)", argIndex))
			args = append(args, textSearch.Text)
			argIndex++
		} else {
			// Regular text search
			escapedText := strings.ReplaceAll(textSearch.Text, "%", "\\%")
			escapedText = strings.ReplaceAll(escapedText, "_", "\\_")
			conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_each_text(n.fields_json) WHERE value ILIKE $%d)", argIndex))
			args = append(args, "%"+escapedText+"%")
			argIndex++
		}
	}

	// State filters (is:marked) - need to check note.marked
	for _, state := range query.States {
		if state == "marked" {
			conditions = append(conditions, "n.marked = TRUE")
		}
	}

	// Build final query
	baseQuery := `
		SELECT DISTINCT n.id, n.user_id, n.guid, n.note_type_id, n.fields_json, n.tags, n.marked, n.created_at, n.updated_at, n.deleted_at
		FROM notes n
	`
	
	// Add JOINs if needed for deck filters
	needsCardJoin := len(query.DecksInclude) > 0 || len(query.DecksExclude) > 0
	if needsCardJoin {
		baseQuery = `
			SELECT DISTINCT n.id, n.user_id, n.guid, n.note_type_id, n.fields_json, n.tags, n.marked, n.created_at, n.updated_at, n.deleted_at
			FROM notes n
		`
	}

	whereClause := strings.Join(conditions, " AND ")
	queryStr := baseQuery + " WHERE " + whereClause + " ORDER BY n.created_at DESC"

	// Add pagination
	queryStr += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find notes by advanced search: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// Ensure NoteRepository implements INoteRepository
var _ secondary.INoteRepository = (*NoteRepository)(nil)

