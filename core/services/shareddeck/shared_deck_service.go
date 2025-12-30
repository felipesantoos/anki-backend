package shareddeck

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// SharedDeckService implements ISharedDeckService
type SharedDeckService struct {
	repo secondary.ISharedDeckRepository
}

// NewSharedDeckService creates a new SharedDeckService instance
func NewSharedDeckService(repo secondary.ISharedDeckRepository) primary.ISharedDeckService {
	return &SharedDeckService{
		repo: repo,
	}
}

// Create publishes a deck to the marketplace
func (s *SharedDeckService) Create(ctx context.Context, authorID int64, name string, description *string, category *string, packagePath string, packageSize int64, tags []string) (*shareddeck.SharedDeck, error) {
	now := time.Now()
	sd, err := shareddeck.NewBuilder().
		WithAuthorID(authorID).
		WithName(name).
		WithDescription(description).
		WithCategory(category).
		WithPackagePath(packagePath).
		WithPackageSize(packageSize).
		WithTags(tags).
		WithIsPublic(true).
		WithCreatedAt(now).
		WithUpdatedAt(now).
		Build()
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, authorID, sd); err != nil {
		return nil, err
	}

	return sd, nil
}

// FindByID finds a shared deck by ID
func (s *SharedDeckService) FindByID(ctx context.Context, userID int64, id int64) (*shareddeck.SharedDeck, error) {
	return s.repo.FindByID(ctx, userID, id)
}

// FindAll finds all public shared decks with optional filters
func (s *SharedDeckService) FindAll(ctx context.Context, category *string, tags []string) ([]*shareddeck.SharedDeck, error) {
	if category != nil {
		return s.repo.FindByCategory(ctx, *category, 100, 0)
	}
	return s.repo.FindPublic(ctx, 100, 0)
}

// Update updates an existing shared deck
func (s *SharedDeckService) Update(ctx context.Context, authorID int64, id int64, name string, description *string, category *string, isPublic bool, tags []string) (*shareddeck.SharedDeck, error) {
	existing, err := s.repo.FindByID(ctx, authorID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("shared deck not found")
	}

	if existing.GetAuthorID() != authorID {
		return nil, fmt.Errorf("not authorized to update this shared deck")
	}

	existing.SetName(name)
	existing.SetDescription(description)
	existing.SetCategory(category)
	existing.SetIsPublic(isPublic)
	existing.SetTags(tags)
	existing.SetUpdatedAt(time.Now())

	if err := s.repo.Update(ctx, authorID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes a shared deck from the marketplace (soft delete)
func (s *SharedDeckService) Delete(ctx context.Context, authorID int64, id int64) error {
	existing, err := s.repo.FindByID(ctx, authorID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("shared deck not found")
	}

	if existing.GetAuthorID() != authorID {
		return fmt.Errorf("not authorized to delete this shared deck")
	}

	return s.repo.Delete(ctx, authorID, id)
}

// IncrementDownloadCount increments the download counter for a shared deck
func (s *SharedDeckService) IncrementDownloadCount(ctx context.Context, userID int64, id int64) error {
	existing, err := s.repo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("shared deck not found")
	}

	existing.SetDownloadCount(existing.GetDownloadCount() + 1)
	return s.repo.Update(ctx, existing.GetAuthorID(), id, existing)
}

