package mappers

import (
	"database/sql"

	savedsearch "github.com/felipesantos/anki-backend/core/domain/entities/saved_search"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// SavedSearchToDomain converts a SavedSearchModel (database representation) to a SavedSearch entity (domain representation)
func SavedSearchToDomain(model *models.SavedSearchModel) (*savedsearch.SavedSearch, error) {
	if model == nil {
		return nil, nil
	}

	builder := savedsearch.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithName(model.Name).
		WithSearchQuery(model.SearchQuery).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// SavedSearchToModel converts a SavedSearch entity (domain representation) to a SavedSearchModel (database representation)
func SavedSearchToModel(savedSearchEntity *savedsearch.SavedSearch) *models.SavedSearchModel {
	model := &models.SavedSearchModel{
		ID:          savedSearchEntity.GetID(),
		UserID:      savedSearchEntity.GetUserID(),
		Name:        savedSearchEntity.GetName(),
		SearchQuery: savedSearchEntity.GetSearchQuery(),
		CreatedAt:   savedSearchEntity.GetCreatedAt(),
		UpdatedAt:   savedSearchEntity.GetUpdatedAt(),
	}

	// Handle nullable deleted_at
	if savedSearchEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *savedSearchEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

