package mappers

import (
	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// DeletionLogToDomain converts a DeletionLogModel (database representation) to a DeletionLog entity (domain representation)
func DeletionLogToDomain(model *models.DeletionLogModel) (*deletionlog.DeletionLog, error) {
	if model == nil {
		return nil, nil
	}

	builder := deletionlog.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithObjectType(model.ObjectType).
		WithObjectID(model.ObjectID).
		WithObjectData(model.ObjectData).
		WithDeletedAt(model.DeletedAt)

	return builder.Build()
}

// DeletionLogToModel converts a DeletionLog entity (domain representation) to a DeletionLogModel (database representation)
func DeletionLogToModel(deletionLogEntity *deletionlog.DeletionLog) *models.DeletionLogModel {
	return &models.DeletionLogModel{
		ID:         deletionLogEntity.GetID(),
		UserID:     deletionLogEntity.GetUserID(),
		ObjectType: deletionLogEntity.GetObjectType(),
		ObjectID:   deletionLogEntity.GetObjectID(),
		ObjectData: deletionLogEntity.GetObjectData(),
		DeletedAt:  deletionLogEntity.GetDeletedAt(),
	}
}

