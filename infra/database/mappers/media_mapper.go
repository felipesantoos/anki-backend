package mappers

import (
	"database/sql"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/infra/database/models"
)

// MediaToDomain converts a MediaModel (database representation) to a Media entity (domain representation)
func MediaToDomain(model *models.MediaModel) (*media.Media, error) {
	if model == nil {
		return nil, nil
	}

	builder := media.NewBuilder().
		WithID(model.ID).
		WithUserID(model.UserID).
		WithFilename(model.Filename).
		WithHash(model.Hash).
		WithSize(model.Size).
		WithMimeType(model.MimeType).
		WithStoragePath(model.StoragePath).
		WithCreatedAt(model.CreatedAt)

	// Handle nullable deleted_at
	if model.DeletedAt.Valid {
		builder = builder.WithDeletedAt(&model.DeletedAt.Time)
	}

	return builder.Build()
}

// MediaToModel converts a Media entity (domain representation) to a MediaModel (database representation)
func MediaToModel(mediaEntity *media.Media) *models.MediaModel {
	model := &models.MediaModel{
		ID:          mediaEntity.GetID(),
		UserID:      mediaEntity.GetUserID(),
		Filename:    mediaEntity.GetFilename(),
		Hash:        mediaEntity.GetHash(),
		Size:        mediaEntity.GetSize(),
		MimeType:    mediaEntity.GetMimeType(),
		StoragePath: mediaEntity.GetStoragePath(),
		CreatedAt:   mediaEntity.GetCreatedAt(),
	}

	// Handle nullable deleted_at
	if mediaEntity.GetDeletedAt() != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *mediaEntity.GetDeletedAt(),
			Valid: true,
		}
	}

	return model
}

