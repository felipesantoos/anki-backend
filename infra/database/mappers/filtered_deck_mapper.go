package mappers

import (
	"database/sql"

	filtereddeck "github.com/felipesantos/anki-backend/core/domain/entities/filtered_deck"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// FilteredDeckToDomain converts a FilteredDeckModel (database representation) to a FilteredDeck entity (domain representation)
func FilteredDeckToDomain(model *models.FilteredDeckModel) (*filtereddeck.FilteredDeck, error) {
	if model == nil {
		return nil, nil
	}

	builder := filtereddeck.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithName(model.Name).
		WithSearchFilter(model.SearchFilter).
		WithLimitCards(model.LimitCards).
		WithOrderBy(model.OrderBy).
		WithReschedule(model.Reschedule).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable second_filter
	if model.SecondFilter.Valid {
		builder = builder.WithSecondFilter(&model.SecondFilter.String)
	}

	// Handle nullable last_rebuild_at
	if model.LastRebuildAt.Valid {
		builder = builder.WithLastRebuildAt(&model.LastRebuildAt.Time)
	}

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// FilteredDeckToModel converts a FilteredDeck entity (domain representation) to a FilteredDeckModel (database representation)
func FilteredDeckToModel(filteredDeckEntity *filtereddeck.FilteredDeck) *models.FilteredDeckModel {
	model := &models.FilteredDeckModel{
		ID:           filteredDeckEntity.GetID(),
		UserID:       filteredDeckEntity.GetUserID(),
		Name:         filteredDeckEntity.GetName(),
		SearchFilter: filteredDeckEntity.GetSearchFilter(),
		LimitCards:   filteredDeckEntity.GetLimitCards(),
		OrderBy:      filteredDeckEntity.GetOrderBy(),
		Reschedule:   filteredDeckEntity.GetReschedule(),
		CreatedAt:    filteredDeckEntity.GetCreatedAt(),
		UpdatedAt:    filteredDeckEntity.GetUpdatedAt(),
	}

	// Handle nullable second_filter
	if filteredDeckEntity.GetSecondFilter() != nil {
		model.SecondFilter = sql.NullString{
			String: *filteredDeckEntity.GetSecondFilter(),
			Valid:  true,
		}
	}

	// Handle nullable last_rebuild_at
	if filteredDeckEntity.GetLastRebuildAt() != nil {
		model.LastRebuildAt = sql.NullTime{
			Time:  *filteredDeckEntity.GetLastRebuildAt(),
			Valid: true,
		}
	}

	// Handle nullable deleted_at
	if filteredDeckEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *filteredDeckEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

