package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	shareddeck "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// SharedDeckRepository implements ISharedDeckRepository using PostgreSQL
type SharedDeckRepository struct {
	db *sql.DB
}

// NewSharedDeckRepository creates a new SharedDeckRepository instance
func NewSharedDeckRepository(db *sql.DB) secondary.ISharedDeckRepository {
	return &SharedDeckRepository{
		db: db,
	}
}

// Save saves or updates a shared deck in the database
func (r *SharedDeckRepository) Save(ctx context.Context, authorID int64, sharedDeckEntity *shareddeck.SharedDeck) error {
	model := mappers.SharedDeckToModel(sharedDeckEntity)

	if sharedDeckEntity.GetID() == 0 {
		// Insert new shared deck
		query := `
			INSERT INTO shared_decks (author_id, name, description, category, package_path, package_size, download_count,
				rating_average, rating_count, tags, is_featured, is_public, created_at, updated_at, deleted_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::TEXT[], $11, $12, $13, $14, $15)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var description interface{}
		if model.Description.Valid {
			description = model.Description.String
		}

		var category interface{}
		if model.Category.Valid {
			category = model.Category.String
		}

		var deletedAt interface{}
		if model.DeletedAt.Valid {
			deletedAt = model.DeletedAt.Time
		}

		var sharedDeckID int64
		tags := sharedDeckEntity.GetTags()
		if tags == nil {
			tags = []string{}
		}
		err := r.db.QueryRowContext(ctx, query,
			authorID,
			model.Name,
			description,
			category,
			model.PackagePath,
			model.PackageSize,
			model.DownloadCount,
			model.RatingAverage,
			model.RatingCount,
			pq.Array(tags),
			model.IsFeatured,
			model.IsPublic,
			model.CreatedAt,
			model.UpdatedAt,
			deletedAt,
		).Scan(&sharedDeckID)
		if err != nil {
			return fmt.Errorf("failed to create shared deck: %w", err)
		}

		sharedDeckEntity.SetID(sharedDeckID)
		return nil
	}

	// Update existing shared deck - validate ownership first
	existingSharedDeck, err := r.FindByID(ctx, authorID, sharedDeckEntity.GetID())
	if err != nil {
		return err
	}
	if existingSharedDeck == nil {
		return ownership.ErrResourceNotFound
	}

	// Update shared deck
	query := `
		UPDATE shared_decks
		SET name = $1, description = $2, category = $3, package_path = $4, package_size = $5, download_count = $6,
			rating_average = $7, rating_count = $8, tags = $9::TEXT[], is_featured = $10, is_public = $11,
			updated_at = $12, deleted_at = $13
		WHERE id = $14 AND author_id = $15 AND deleted_at IS NULL
	`

	now := time.Now()
	model.UpdatedAt = now

	var description interface{}
	if model.Description.Valid {
		description = model.Description.String
	}

	var category interface{}
	if model.Category.Valid {
		category = model.Category.String
	}

	var deletedAt interface{}
	if model.DeletedAt.Valid {
		deletedAt = model.DeletedAt.Time
	}

	tags := sharedDeckEntity.GetTags()
	if tags == nil {
		tags = []string{}
	}

	result, err := r.db.ExecContext(ctx, query,
		model.Name,
		description,
		category,
		model.PackagePath,
		model.PackageSize,
		model.DownloadCount,
		model.RatingAverage,
		model.RatingCount,
		pq.Array(tags),
		model.IsFeatured,
		model.IsPublic,
		model.UpdatedAt,
		deletedAt,
		model.ID,
		authorID,
	)

	if err != nil {
		return fmt.Errorf("failed to update shared deck: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// FindByID finds a shared deck by ID
// For non-authors, only returns public shared decks
// For authors, returns their own shared decks regardless of public status
func (r *SharedDeckRepository) FindByID(ctx context.Context, userID int64, id int64) (*shareddeck.SharedDeck, error) {
	query := `
		SELECT id, author_id, name, description, category, package_path, package_size, download_count,
			rating_average, rating_count, tags, is_featured, is_public, created_at, updated_at, deleted_at
		FROM shared_decks
		WHERE id = $1 AND deleted_at IS NULL AND (is_public = TRUE OR author_id = $2)
	`

	var model models.SharedDeckModel
	var description sql.NullString
	var category sql.NullString
	var tags pq.StringArray
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.AuthorID,
		&model.Name,
		&description,
		&category,
		&model.PackagePath,
		&model.PackageSize,
		&model.DownloadCount,
		&model.RatingAverage,
		&model.RatingCount,
		&tags,
		&model.IsFeatured,
		&model.IsPublic,
		&model.CreatedAt,
		&model.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find shared deck: %w", err)
	}

	model.Description = description
	model.Category = category
	if len(tags) > 0 {
		model.Tags = sql.NullString{String: "{" + strings.Join(tags, ",") + "}", Valid: true}
	}
	if deletedAt.Valid {
		model.DeletedAt = deletedAt
	}

	return mappers.SharedDeckToDomain(&model)
}

// FindByAuthorID finds all shared decks by an author
func (r *SharedDeckRepository) FindByAuthorID(ctx context.Context, authorID int64) ([]*shareddeck.SharedDeck, error) {
	query := `
		SELECT id, author_id, name, description, category, package_path, package_size, download_count,
			rating_average, rating_count, tags, is_featured, is_public, created_at, updated_at, deleted_at
		FROM shared_decks
		WHERE author_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shared decks by author ID: %w", err)
	}
	defer rows.Close()

	var sharedDecks []*shareddeck.SharedDeck
	for rows.Next() {
		var model models.SharedDeckModel
		var description sql.NullString
		var category sql.NullString
		var tags pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.AuthorID,
			&model.Name,
			&description,
			&category,
			&model.PackagePath,
			&model.PackageSize,
			&model.DownloadCount,
			&model.RatingAverage,
			&model.RatingCount,
			&tags,
			&model.IsFeatured,
			&model.IsPublic,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared deck: %w", err)
		}

		model.Description = description
		model.Category = category
		if len(tags) > 0 {
			model.Tags = sql.NullString{String: "{" + strings.Join(tags, ",") + "}", Valid: true}
		}
		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		sharedDeckEntity, err := mappers.SharedDeckToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert shared deck to domain: %w", err)
		}
		sharedDecks = append(sharedDecks, sharedDeckEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shared decks: %w", err)
	}

	return sharedDecks, nil
}

// Update updates an existing shared deck, validating ownership
func (r *SharedDeckRepository) Update(ctx context.Context, authorID int64, id int64, sharedDeckEntity *shareddeck.SharedDeck) error {
	return r.Save(ctx, authorID, sharedDeckEntity)
}

// Delete deletes a shared deck, validating ownership (soft delete)
func (r *SharedDeckRepository) Delete(ctx context.Context, authorID int64, id int64) error {
	// Validate ownership first
	existingSharedDeck, err := r.FindByID(ctx, authorID, id)
	if err != nil {
		return err
	}
	if existingSharedDeck == nil || existingSharedDeck.GetAuthorID() != authorID {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE shared_decks
		SET deleted_at = $1
		WHERE id = $2 AND author_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id, authorID)
	if err != nil {
		return fmt.Errorf("failed to delete shared deck: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ownership.ErrResourceNotFound
	}

	return nil
}

// Exists checks if a shared deck exists
func (r *SharedDeckRepository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM shared_decks
			WHERE id = $1 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check shared deck existence: %w", err)
	}

	return exists, nil
}

// FindPublic finds all public shared decks (visible to all users)
func (r *SharedDeckRepository) FindPublic(ctx context.Context, limit, offset int) ([]*shareddeck.SharedDeck, error) {
	query := `
		SELECT id, author_id, name, description, category, package_path, package_size, download_count,
			rating_average, rating_count, tags, is_featured, is_public, created_at, updated_at, deleted_at
		FROM shared_decks
		WHERE is_public = TRUE AND deleted_at IS NULL
		ORDER BY rating_average DESC, download_count DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find public shared decks: %w", err)
	}
	defer rows.Close()

	var sharedDecks []*shareddeck.SharedDeck
	for rows.Next() {
		var model models.SharedDeckModel
		var description sql.NullString
		var category sql.NullString
		var tags pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.AuthorID,
			&model.Name,
			&description,
			&category,
			&model.PackagePath,
			&model.PackageSize,
			&model.DownloadCount,
			&model.RatingAverage,
			&model.RatingCount,
			&tags,
			&model.IsFeatured,
			&model.IsPublic,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared deck: %w", err)
		}

		model.Description = description
		model.Category = category
		if len(tags) > 0 {
			model.Tags = sql.NullString{String: "{" + strings.Join(tags, ",") + "}", Valid: true}
		}
		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		sharedDeckEntity, err := mappers.SharedDeckToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert shared deck to domain: %w", err)
		}
		sharedDecks = append(sharedDecks, sharedDeckEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shared decks: %w", err)
	}

	return sharedDecks, nil
}

// FindByCategory finds public shared decks by category
func (r *SharedDeckRepository) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*shareddeck.SharedDeck, error) {
	query := `
		SELECT id, author_id, name, description, category, package_path, package_size, download_count,
			rating_average, rating_count, tags, is_featured, is_public, created_at, updated_at, deleted_at
		FROM shared_decks
		WHERE is_public = TRUE AND category = $1 AND deleted_at IS NULL
		ORDER BY rating_average DESC, download_count DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find shared decks by category: %w", err)
	}
	defer rows.Close()

	var sharedDecks []*shareddeck.SharedDeck
	for rows.Next() {
		var model models.SharedDeckModel
		var description sql.NullString
		var category sql.NullString
		var tags pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.AuthorID,
			&model.Name,
			&description,
			&category,
			&model.PackagePath,
			&model.PackageSize,
			&model.DownloadCount,
			&model.RatingAverage,
			&model.RatingCount,
			&tags,
			&model.IsFeatured,
			&model.IsPublic,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared deck: %w", err)
		}

		model.Description = description
		model.Category = category
		if len(tags) > 0 {
			model.Tags = sql.NullString{String: "{" + strings.Join(tags, ",") + "}", Valid: true}
		}
		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		sharedDeckEntity, err := mappers.SharedDeckToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert shared deck to domain: %w", err)
		}
		sharedDecks = append(sharedDecks, sharedDeckEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shared decks: %w", err)
	}

	return sharedDecks, nil
}

// FindFeatured finds featured public shared decks
func (r *SharedDeckRepository) FindFeatured(ctx context.Context, limit int) ([]*shareddeck.SharedDeck, error) {
	query := `
		SELECT id, author_id, name, description, category, package_path, package_size, download_count,
			rating_average, rating_count, tags, is_featured, is_public, created_at, updated_at, deleted_at
		FROM shared_decks
		WHERE is_public = TRUE AND is_featured = TRUE AND deleted_at IS NULL
		ORDER BY rating_average DESC, download_count DESC, created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find featured shared decks: %w", err)
	}
	defer rows.Close()

	var sharedDecks []*shareddeck.SharedDeck
	for rows.Next() {
		var model models.SharedDeckModel
		var description sql.NullString
		var category sql.NullString
		var tags pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.AuthorID,
			&model.Name,
			&description,
			&category,
			&model.PackagePath,
			&model.PackageSize,
			&model.DownloadCount,
			&model.RatingAverage,
			&model.RatingCount,
			&tags,
			&model.IsFeatured,
			&model.IsPublic,
			&model.CreatedAt,
			&model.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared deck: %w", err)
		}

		model.Description = description
		model.Category = category
		if len(tags) > 0 {
			model.Tags = sql.NullString{String: "{" + strings.Join(tags, ",") + "}", Valid: true}
		}
		if deletedAt.Valid {
			model.DeletedAt = deletedAt
		}

		sharedDeckEntity, err := mappers.SharedDeckToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert shared deck to domain: %w", err)
		}
		sharedDecks = append(sharedDecks, sharedDeckEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shared decks: %w", err)
	}

	return sharedDecks, nil
}

// Ensure SharedDeckRepository implements ISharedDeckRepository
var _ secondary.ISharedDeckRepository = (*SharedDeckRepository)(nil)

