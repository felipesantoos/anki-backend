package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
)

// ToSyncMetaResponse converts SyncMeta domain entity to Response DTO
func ToSyncMetaResponse(sm *syncmeta.SyncMeta) *response.SyncMetaResponse {
	if sm == nil {
		return nil
	}
	return &response.SyncMetaResponse{
		ID:          sm.GetID(),
		UserID:      sm.GetUserID(),
		ClientID:    sm.GetClientID(),
		LastSync:    sm.GetLastSync(),
		LastSyncUSN: sm.GetLastSyncUSN(),
		CreatedAt:   sm.GetCreatedAt(),
		UpdatedAt:   sm.GetUpdatedAt(),
	}
}

