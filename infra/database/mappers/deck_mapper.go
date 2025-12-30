package mappers

import (
	"database/sql"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// DeckToDomain converts a DeckModel (database representation) to a Deck entity (domain representation)
func DeckToDomain(model *models.DeckModel) (*deck.Deck, error) {
	if model == nil {
		return nil, nil
	}

	deckEntity := &deck.Deck{}
	deckEntity.SetID(model.ID)
	deckEntity.SetUserID(model.UserID)
	deckEntity.SetName(model.Name)
	deckEntity.SetOptionsJSON(model.OptionsJSON)
	deckEntity.SetCreatedAt(model.CreatedAt)
	deckEntity.SetUpdatedAt(model.UpdatedAt)

	if model.ParentID.Valid {
		parentID := model.ParentID.Int64
		deckEntity.SetParentID(&parentID)
	}

	if model.DeletedAt.Valid {
		deletedAt := model.DeletedAt.Time
		deckEntity.SetDeletedAt(&deletedAt)
	}

	return deckEntity, nil
}

// DeckToModel converts a Deck entity (domain representation) to a DeckModel (database representation)
func DeckToModel(deckEntity *deck.Deck) *models.DeckModel {
	if deckEntity == nil {
		return nil
	}

	model := &models.DeckModel{
		ID:          deckEntity.GetID(),
		UserID:      deckEntity.GetUserID(),
		Name:        deckEntity.GetName(),
		OptionsJSON: deckEntity.GetOptionsJSON(),
		CreatedAt:   deckEntity.GetCreatedAt(),
		UpdatedAt:   deckEntity.GetUpdatedAt(),
	}

	if deckEntity.GetParentID() != nil {
		model.ParentID = sql.NullInt64{
			Int64: *deckEntity.GetParentID(),
			Valid: true,
		}
	}

	if deckEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *deckEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

