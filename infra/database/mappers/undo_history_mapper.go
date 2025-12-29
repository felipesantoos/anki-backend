package mappers

import (
	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// UndoHistoryToDomain converts an UndoHistoryModel (database representation) to an UndoHistory entity (domain representation)
func UndoHistoryToDomain(model *models.UndoHistoryModel) (*undohistory.UndoHistory, error) {
	if model == nil {
		return nil, nil
	}

	builder := undohistory.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithOperationType(model.OperationType).
		WithOperationData(model.OperationData).
		WithCreatedAt(model.CreatedAt)

	return builder.Build()
}

// UndoHistoryToModel converts an UndoHistory entity (domain representation) to an UndoHistoryModel (database representation)
func UndoHistoryToModel(undoHistoryEntity *undohistory.UndoHistory) *models.UndoHistoryModel {
	return &models.UndoHistoryModel{
		ID:            undoHistoryEntity.GetID(),
		UserID:        undoHistoryEntity.GetUserID(),
		OperationType: undoHistoryEntity.GetOperationType(),
		OperationData: undoHistoryEntity.GetOperationData(),
		CreatedAt:     undoHistoryEntity.GetCreatedAt(),
	}
}

