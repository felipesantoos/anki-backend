package mappers

import (
	"database/sql"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// NoteTypeToDomain converts a NoteTypeModel (database representation) to a NoteType entity (domain representation)
func NoteTypeToDomain(model *models.NoteTypeModel) (*notetype.NoteType, error) {
	if model == nil {
		return nil, nil
	}

	builder := notetype.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithName(model.Name).
		WithFieldsJSON(model.FieldsJSON).
		WithCardTypesJSON(model.CardTypesJSON).
		WithTemplatesJSON(model.TemplatesJSON).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// NoteTypeToModel converts a NoteType entity (domain representation) to a NoteTypeModel (database representation)
func NoteTypeToModel(noteTypeEntity *notetype.NoteType) *models.NoteTypeModel {
	model := &models.NoteTypeModel{
		ID:            noteTypeEntity.GetID(),
		UserID:        noteTypeEntity.GetUserID(),
		Name:          noteTypeEntity.GetName(),
		FieldsJSON:    noteTypeEntity.GetFieldsJSON(),
		CardTypesJSON: noteTypeEntity.GetCardTypesJSON(),
		TemplatesJSON: noteTypeEntity.GetTemplatesJSON(),
		CreatedAt:     noteTypeEntity.GetCreatedAt(),
		UpdatedAt:     noteTypeEntity.GetUpdatedAt(),
	}

	// Handle nullable deleted_at
	if noteTypeEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *noteTypeEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

