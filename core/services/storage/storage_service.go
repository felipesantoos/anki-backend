package storage

import (
	"fmt"
	"log/slog"

	"github.com/felipesantos/anki-backend/config"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	localStorage "github.com/felipesantos/anki-backend/infra/storage/local"
	s3Storage "github.com/felipesantos/anki-backend/infra/storage/s3"
	cloudflareStorage "github.com/felipesantos/anki-backend/infra/storage/cloudflare"
)

// StorageService provides high-level file storage operations
type StorageService struct {
	repository secondary.IStorageRepository
	logger     *slog.Logger
}

// NewStorageService creates a new storage service with the appropriate repository
func NewStorageService(repository secondary.IStorageRepository, logger *slog.Logger) *StorageService {
	return &StorageService{
		repository: repository,
		logger:     logger,
	}
}

// NewStorageRepository creates a storage repository based on configuration
func NewStorageRepository(cfg config.StorageConfig, logger *slog.Logger) (secondary.IStorageRepository, error) {
	switch cfg.Type {
	case "local":
		repo, err := localStorage.NewLocalStorageRepository(cfg.LocalPath, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create local storage repository: %w", err)
		}
		return repo, nil

	case "s3":
		repo, err := s3Storage.NewS3StorageRepository(
			cfg.S3Bucket,
			cfg.S3Region,
			cfg.S3Key,
			cfg.S3Secret,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 storage repository: %w", err)
		}
		return repo, nil

	case "cloudflare", "r2":
		repo, err := cloudflareStorage.NewCloudflareR2StorageRepository(
			cfg.CloudflareAccountID,
			cfg.CloudflareR2Bucket,
			cfg.CloudflareR2Key,
			cfg.CloudflareR2Secret,
			cfg.CloudflareR2Endpoint,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloudflare R2 storage repository: %w", err)
		}
		return repo, nil

	default:
		return nil, fmt.Errorf("unsupported storage type: %s (supported: local, s3, cloudflare, r2)", cfg.Type)
	}
}

// Repository returns the underlying storage repository
func (s *StorageService) Repository() secondary.IStorageRepository {
	return s.repository
}

