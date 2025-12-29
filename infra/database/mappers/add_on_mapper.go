package mappers

import (
	addon "github.com/felipesantos/anki-backend/core/domain/entities/add_on"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// AddOnToDomain converts an AddOnModel (database representation) to an AddOn entity (domain representation)
func AddOnToDomain(model *models.AddOnModel) (*addon.AddOn, error) {
	if model == nil {
		return nil, nil
	}

	builder := addon.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithCode(model.Code).
		WithName(model.Name).
		WithVersion(model.Version).
		WithEnabled(model.Enabled).
		WithConfigJSON(model.ConfigJSON).
		WithInstalledAt(model.InstalledAt).
		WithUpdatedAt(model.UpdatedAt)

	return builder.Build()
}

// AddOnToModel converts an AddOn entity (domain representation) to an AddOnModel (database representation)
func AddOnToModel(addOnEntity *addon.AddOn) *models.AddOnModel {
	return &models.AddOnModel{
		ID:          addOnEntity.GetID(),
		UserID:      addOnEntity.GetUserID(),
		Code:        addOnEntity.GetCode(),
		Name:        addOnEntity.GetName(),
		Version:     addOnEntity.GetVersion(),
		Enabled:     addOnEntity.GetEnabled(),
		ConfigJSON:  addOnEntity.GetConfigJSON(),
		InstalledAt: addOnEntity.GetInstalledAt(),
		UpdatedAt:   addOnEntity.GetUpdatedAt(),
	}
}

