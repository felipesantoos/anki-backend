package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	deletionlog "github.com/felipesantos/anki-backend/core/domain/entities/deletion_log"
)

// ToDeletionLogResponse converts DeletionLog domain entity to Response DTO
func ToDeletionLogResponse(dl *deletionlog.DeletionLog) *response.DeletionLogResponse {
	if dl == nil {
		return nil
	}
	return &response.DeletionLogResponse{
		ID:         dl.GetID(),
		UserID:     dl.GetUserID(),
		ObjectType: dl.GetObjectType(),
		ObjectID:   dl.GetObjectID(),
		DeletedAt:  dl.GetDeletedAt(),
	}
}

// ToDeletionLogResponseList converts list of DeletionLog domain entities to list of Response DTOs
func ToDeletionLogResponseList(logs []*deletionlog.DeletionLog) []*response.DeletionLogResponse {
	res := make([]*response.DeletionLogResponse, len(logs))
	for i, dl := range logs {
		res[i] = ToDeletionLogResponse(dl)
	}
	return res
}

