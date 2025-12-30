package media

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// MediaService implements IMediaService
type MediaService struct {
	repo secondary.IMediaRepository
}

// NewMediaService creates a new MediaService instance
func NewMediaService(repo secondary.IMediaRepository) primary.IMediaService {
	return &MediaService{
		repo: repo,
	}
}

// Create records a new media file
func (s *MediaService) Create(ctx context.Context, userID int64, filename string, hash string, size int64, mimeType string, storagePath string) (*media.Media, error) {
	now := time.Now()
	m, err := media.NewBuilder().
		WithUserID(userID).
		WithFilename(filename).
		WithHash(hash).
		WithSize(size).
		WithMimeType(mimeType).
		WithStoragePath(storagePath).
		WithCreatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, userID, m); err != nil {
		return nil, err
	}

	return m, nil
}

// FindByID finds a media file by ID
func (s *MediaService) FindByID(ctx context.Context, userID int64, id int64) (*media.Media, error) {
	return s.repo.FindByID(ctx, userID, id)
}

// FindByUserID finds all media files for a user
func (s *MediaService) FindByUserID(ctx context.Context, userID int64) ([]*media.Media, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// Delete removes a media file record (soft delete)
func (s *MediaService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

