package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	undohistory "github.com/felipesantos/anki-backend/core/domain/entities/undo_history"
)

// ToUndoHistoryResponse converts UndoHistory domain entity to Response DTO
func ToUndoHistoryResponse(uh *undohistory.UndoHistory) *response.UndoHistoryResponse {
	if uh == nil {
		return nil
	}
	return &response.UndoHistoryResponse{
		ID:            uh.GetID(),
		UserID:        uh.GetUserID(),
		OperationType: uh.GetOperationType(),
		OperationData: uh.GetOperationData(),
		CreatedAt:     uh.GetCreatedAt(),
	}
}

// ToUndoHistoryResponseList converts list of UndoHistory domain entities to list of Response DTOs
func ToUndoHistoryResponseList(history []*undohistory.UndoHistory) []*response.UndoHistoryResponse {
	res := make([]*response.UndoHistoryResponse, len(history))
	for i, uh := range history {
		res[i] = ToUndoHistoryResponse(uh)
	}
	return res
}

