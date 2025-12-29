package mappers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// NoteToDomain converts a NoteModel (database representation) to a Note entity (domain representation)
func NoteToDomain(model *models.NoteModel) (*note.Note, error) {
	if model == nil {
		return nil, nil
	}

	// Create GUID value object
	guid, err := valueobjects.NewGUID(model.GUID)
	if err != nil {
		return nil, fmt.Errorf("failed to create GUID: %w", err)
	}

	builder := note.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithGUID(guid).
		WithNoteTypeID(model.NoteTypeID).
		WithFieldsJSON(model.FieldsJSON).
		WithMarked(model.Marked).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle tags (TEXT[] array)
	// PostgreSQL TEXT[] is returned as a string like "{tag1,tag2,tag3}"
	if model.Tags.Valid && model.Tags.String != "" {
		tagsStr := model.Tags.String
		var tags []string
		
		// Parse PostgreSQL array format: {tag1,tag2,tag3}
		if len(tagsStr) >= 2 && tagsStr[0] == '{' && tagsStr[len(tagsStr)-1] == '}' {
			// Remove braces
			inner := tagsStr[1 : len(tagsStr)-1]
			if inner != "" {
				// Split by comma (simplified - doesn't handle quoted strings with commas)
				// This works for most cases since tags usually don't contain commas
				parts := strings.Split(inner, ",")
				tags = make([]string, 0, len(parts))
				for _, part := range parts {
					part = strings.TrimSpace(part)
					// Remove quotes if present
					if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
						part = part[1 : len(part)-1]
					}
					if part != "" {
						tags = append(tags, part)
					}
				}
			}
		}
		builder = builder.WithTags(tags)
	}

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// NoteToModel converts a Note entity (domain representation) to a NoteModel (database representation)
func NoteToModel(noteEntity *note.Note) *models.NoteModel {
	model := &models.NoteModel{
		ID:         noteEntity.GetID(),
		UserID:     noteEntity.GetUserID(),
		GUID:       noteEntity.GetGUID().Value(),
		NoteTypeID: noteEntity.GetNoteTypeID(),
		FieldsJSON: noteEntity.GetFieldsJSON(),
		Tags:       sql.NullString{},
		Marked:     noteEntity.GetMarked(),
		CreatedAt:  noteEntity.GetCreatedAt(),
		UpdatedAt:  noteEntity.GetUpdatedAt(),
	}

	// Handle tags - convert slice to PostgreSQL TEXT[] format string
	// PostgreSQL TEXT[] format: {tag1,tag2,tag3}
	tags := noteEntity.GetTags()
	if len(tags) > 0 {
		// Build PostgreSQL array format string
		tagsStr := "{"
		for i, tag := range tags {
			if i > 0 {
				tagsStr += ","
			}
			// Escape quotes and backslashes if present
			escapedTag := strings.ReplaceAll(tag, "\\", "\\\\")
			escapedTag = strings.ReplaceAll(escapedTag, "\"", "\\\"")
			// Quote tag if it contains comma, space, or special characters
			if strings.ContainsAny(escapedTag, ", \t\n{}") {
				tagsStr += `"` + escapedTag + `"`
			} else {
				tagsStr += escapedTag
			}
		}
		tagsStr += "}"
		model.Tags = sql.NullString{
			String: tagsStr,
			Valid:  true,
		}
	}

	// Handle nullable deleted_at
	if noteEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *noteEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

