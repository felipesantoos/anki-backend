package mappers

import (
	"database/sql"

	deckoptionspreset "github.com/felipesantos/anki-backend/core/domain/entities/deck_options_preset"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// DeckOptionsPresetToDomain converts a DeckOptionsPresetModel (database representation) to a DeckOptionsPreset entity (domain representation)
func DeckOptionsPresetToDomain(model *models.DeckOptionsPresetModel) (*deckoptionspreset.DeckOptionsPreset, error) {
	if model == nil {
		return nil, nil
	}

	builder := deckoptionspreset.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithName(model.Name).
		WithOptionsJSON(model.OptionsJSON).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// DeckOptionsPresetToModel converts a DeckOptionsPreset entity (domain representation) to a DeckOptionsPresetModel (database representation)
func DeckOptionsPresetToModel(presetEntity *deckoptionspreset.DeckOptionsPreset) *models.DeckOptionsPresetModel {
	model := &models.DeckOptionsPresetModel{
		ID:          presetEntity.GetID(),
		UserID:      presetEntity.GetUserID(),
		Name:        presetEntity.GetName(),
		OptionsJSON: presetEntity.GetOptionsJSON(),
		CreatedAt:   presetEntity.GetCreatedAt(),
		UpdatedAt:   presetEntity.GetUpdatedAt(),
	}

	// Handle nullable deleted_at
	if presetEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *presetEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

