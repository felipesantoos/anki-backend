package mappers

import (
	syncmeta "github.com/felipesantos/anki-backend/core/domain/entities/sync_meta"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// SyncMetaToDomain converts a SyncMetaModel (database representation) to a SyncMeta entity (domain representation)
func SyncMetaToDomain(model *models.SyncMetaModel) (*syncmeta.SyncMeta, error) {
	if model == nil {
		return nil, nil
	}

	builder := syncmeta.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithClientID(model.ClientID).
		WithLastSync(model.LastSync).
		WithLastSyncUSN(model.LastSyncUSN).
		WithCreatedAt(model.CreatedAt).
		WithUpdatedAt(model.UpdatedAt)

	return builder.Build()
}

// SyncMetaToModel converts a SyncMeta entity (domain representation) to a SyncMetaModel (database representation)
func SyncMetaToModel(syncMetaEntity *syncmeta.SyncMeta) *models.SyncMetaModel {
	return &models.SyncMetaModel{
		ID:          syncMetaEntity.GetID(),
		UserID:      syncMetaEntity.GetUserID(),
		ClientID:    syncMetaEntity.GetClientID(),
		LastSync:    syncMetaEntity.GetLastSync(),
		LastSyncUSN: syncMetaEntity.GetLastSyncUSN(),
		CreatedAt:   syncMetaEntity.GetCreatedAt(),
		UpdatedAt:   syncMetaEntity.GetUpdatedAt(),
	}
}

