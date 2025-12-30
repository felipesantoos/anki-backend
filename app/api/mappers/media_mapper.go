package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
)

// ToMediaResponse converts Media domain entity to Response DTO
func ToMediaResponse(m *media.Media) *response.MediaResponse {
	if m == nil {
		return nil
	}
	return &response.MediaResponse{
		ID:          m.GetID(),
		UserID:      m.GetUserID(),
		Filename:    m.GetFilename(),
		Hash:        m.GetHash(),
		Size:        m.GetSize(),
		MimeType:    m.GetMimeType(),
		StoragePath: m.GetStoragePath(),
		CreatedAt:   m.GetCreatedAt(),
	}
}

// ToMediaResponseList converts list of Media domain entities to list of Response DTOs
func ToMediaResponseList(mediaFiles []*media.Media) []*response.MediaResponse {
	res := make([]*response.MediaResponse, len(mediaFiles))
	for i, m := range mediaFiles {
		res[i] = ToMediaResponse(m)
	}
	return res
}

