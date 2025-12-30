package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
)

// IMediaService defines the interface for media file management
type IMediaService interface {
	// Create records a new media file
	Create(ctx context.Context, userID int64, filename string, hash string, size int64, mimeType string, storagePath string) (*media.Media, error)

	// FindByID finds a media file by ID
	FindByID(ctx context.Context, userID int64, id int64) (*media.Media, error)

	// FindByUserID finds all media files for a user
	FindByUserID(ctx context.Context, userID int64) ([]*media.Media, error)

	// Delete removes a media file record (soft delete)
	Delete(ctx context.Context, userID int64, id int64) error
}

