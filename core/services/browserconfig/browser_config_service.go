package browserconfig

import (
	"context"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/browser_config"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// BrowserConfigService implements IBrowserConfigService
type BrowserConfigService struct {
	repo secondary.IBrowserConfigRepository
}

// NewBrowserConfigService creates a new BrowserConfigService instance
func NewBrowserConfigService(repo secondary.IBrowserConfigRepository) primary.IBrowserConfigService {
	return &BrowserConfigService{
		repo: repo,
	}
}

// FindByUserID finds browser config for a user
func (s *BrowserConfigService) FindByUserID(ctx context.Context, userID int64) (*browserconfig.BrowserConfig, error) {
	cfg, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		// Create default config if not found
		now := time.Now()
		cfg, err = browserconfig.NewBuilder().
			WithUserID(userID).
			WithVisibleColumns([]string{"note", "deck", "tags", "due", "interval", "ease"}).
			WithColumnWidths("{}").
			WithSortDirection("asc").
			WithCreatedAt(now).
			WithUpdatedAt(now).
			Build()
		if err != nil {
			return nil, err
		}

		if err := s.repo.Save(ctx, userID, cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// Update updates browser config
func (s *BrowserConfigService) Update(ctx context.Context, userID int64, config *browserconfig.BrowserConfig) error {
	config.SetUpdatedAt(time.Now())
	return s.repo.Update(ctx, userID, config.GetID(), config)
}

