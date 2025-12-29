package mappers

import (
	"database/sql"

	"github.com/felipesantos/anki-backend/core/domain/entities/profile"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// ProfileToDomain converts a ProfileModel (database representation) to a Profile entity (domain representation)
func ProfileToDomain(model *models.ProfileModel) (*profile.Profile, error) {
	if model == nil {
		return nil, nil
	}

	builder := profile.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithName(model.Name).
		WithAnkiWebSyncEnabled(model.AnkiWebSyncEnabled).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	// Handle nullable ankiweb_username
	if model.AnkiWebUsername.Valid {
		builder = builder.WithAnkiWebUsername(&model.AnkiWebUsername.String)
	}

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// ProfileToModel converts a Profile entity (domain representation) to a ProfileModel (database representation)
func ProfileToModel(profileEntity *profile.Profile) *models.ProfileModel {
	model := &models.ProfileModel{
		ID:                 profileEntity.GetID(),
		UserID:             profileEntity.GetUserID(),
		Name:               profileEntity.GetName(),
		AnkiWebSyncEnabled: profileEntity.GetAnkiWebSyncEnabled(),
		CreatedAt:          profileEntity.GetCreatedAt(),
		UpdatedAt:          profileEntity.GetUpdatedAt(),
	}

	// Handle nullable ankiweb_username
	if profileEntity.GetAnkiWebUsername() != nil {
		model.AnkiWebUsername = sql.NullString{
			String: *profileEntity.GetAnkiWebUsername(),
			Valid:  true,
		}
	}

	// Handle nullable deleted_at
	if profileEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *profileEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

