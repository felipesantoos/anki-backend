package mappers

import (
	flagname "github.com/felipesantos/anki-backend/core/domain/entities/flag_name"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// FlagNameToDomain converts a FlagNameModel (database representation) to a FlagName entity (domain representation)
func FlagNameToDomain(model *models.FlagNameModel) (*flagname.FlagName, error) {
	if model == nil {
		return nil, nil
	}

	builder := flagname.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithFlagNumber(model.FlagNumber).
		WithName(model.Name).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	return builder.Build()
}

// FlagNameToModel converts a FlagName entity (domain representation) to a FlagNameModel (database representation)
func FlagNameToModel(flagNameEntity *flagname.FlagName) *models.FlagNameModel {
	return &models.FlagNameModel{
		ID:         flagNameEntity.GetID(),
		UserID:     flagNameEntity.GetUserID(),
		FlagNumber: flagNameEntity.GetFlagNumber(),
		Name:       flagNameEntity.GetName(),
		CreatedAt:  flagNameEntity.GetCreatedAt(),
		UpdatedAt:  flagNameEntity.GetUpdatedAt(),
	}
}

