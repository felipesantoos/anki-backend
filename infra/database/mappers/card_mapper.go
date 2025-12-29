package mappers

import (
	"database/sql"
	"fmt"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// CardToDomain converts a CardModel (database representation) to a Card entity (domain representation)
func CardToDomain(model *models.CardModel) (*card.Card, error) {
	if model == nil {
		return nil, nil
	}

	// Parse card state from string
	cardState := valueobjects.CardState(model.State)
	if !cardState.IsValid() {
		return nil, fmt.Errorf("invalid card state: %s", model.State)
	}

	builder := card.NewBuilder().
		WithID(model.ID).
		WithNoteID(model.NoteID).
		WithCardTypeID(model.CardTypeID).
		WithDeckID(model.DeckID).
		WithDue(model.Due).
		WithInterval(model.Interval).
		WithEase(model.Ease).
		WithLapses(model.Lapses).
		WithReps(model.Reps).
		WithState(cardState).
		WithPosition(model.Position).
		WithFlag(model.Flag).
		WithSuspended(model.Suspended).
		WithBuried(model.Buried).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable home_deck_id
	if model.HomeDeckID.Valid {
		builder = builder.WithHomeDeckID(&model.HomeDeckID.Int64)
	}

	// Handle nullable stability
	if model.Stability.Valid {
		builder = builder.WithStability(&model.Stability.Float64)
	}

	// Handle nullable difficulty
	if model.Difficulty.Valid {
		builder = builder.WithDifficulty(&model.Difficulty.Float64)
	}

	// Handle nullable last_review_at
	if model.LastReviewAt.Valid {
		builder = builder.WithLastReviewAt(&model.LastReviewAt.Time)
	}

	return builder.Build()
}

// CardToModel converts a Card entity (domain representation) to a CardModel (database representation)
func CardToModel(cardEntity *card.Card) *models.CardModel {
	model := &models.CardModel{
		ID:         cardEntity.GetID(),
		NoteID:     cardEntity.GetNoteID(),
		CardTypeID: cardEntity.GetCardTypeID(),
		DeckID:     cardEntity.GetDeckID(),
		Due:        cardEntity.GetDue(),
		Interval:   cardEntity.GetInterval(),
		Ease:       cardEntity.GetEase(),
		Lapses:     cardEntity.GetLapses(),
		Reps:       cardEntity.GetReps(),
		State:      cardEntity.GetState().String(),
		Position:   cardEntity.GetPosition(),
		Flag:       cardEntity.GetFlag(),
		Suspended:  cardEntity.GetSuspended(),
		Buried:     cardEntity.GetBuried(),
		CreatedAt:  cardEntity.GetCreatedAt(),
		UpdatedAt:  cardEntity.GetUpdatedAt(),
	}

	// Handle nullable home_deck_id
	if cardEntity.GetHomeDeckID() != nil {
		model.HomeDeckID = sql.NullInt64{
			Int64: *cardEntity.GetHomeDeckID(),
			Valid: true,
		}
	}

	// Handle nullable stability
	if cardEntity.GetStability() != nil {
		model.Stability = sql.NullFloat64{
			Float64: *cardEntity.GetStability(),
			Valid:   true,
		}
	}

	// Handle nullable difficulty
	if cardEntity.GetDifficulty() != nil {
		model.Difficulty = sql.NullFloat64{
			Float64: *cardEntity.GetDifficulty(),
			Valid:   true,
		}
	}

	// Handle nullable last_review_at
	if cardEntity.GetLastReviewAt() != nil {
		model.LastReviewAt = sql.NullTime{
			Time:  *cardEntity.GetLastReviewAt(),
			Valid: true,
		}
	}

	return model
}

