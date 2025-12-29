package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	shareddeckrating "github.com/felipesantos/anki-backend/core/domain/entities/shared_deck_rating"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// SharedDeckRatingRepository implements ISharedDeckRatingRepository using PostgreSQL
type SharedDeckRatingRepository struct {
	db *sql.DB
}

// NewSharedDeckRatingRepository creates a new SharedDeckRatingRepository instance
func NewSharedDeckRatingRepository(db *sql.DB) secondary.ISharedDeckRatingRepository {
	return &SharedDeckRatingRepository{
		db: db,
	}
}

// Save saves or updates a shared deck rating in the database
func (r *SharedDeckRatingRepository) Save(ctx context.Context, userID int64, ratingEntity *shareddeckrating.SharedDeckRating) error {
	model := mappers.SharedDeckRatingToModel(ratingEntity)

	if ratingEntity.GetID() == 0 {
		// Insert new rating
		query := `
			INSERT INTO shared_deck_ratings (user_id, shared_deck_id, rating, comment, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var comment interface{}
		if model.Comment.Valid {
			comment = model.Comment.String
		}

		var ratingID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.SharedDeckID,
			model.Rating,
			comment,
			model.CreatedAt,
			model.UpdatedAt,
		).Scan(&ratingID)
		if err != nil {
			return fmt.Errorf("failed to create shared deck rating: %w", err)
		}

		ratingEntity.SetID(ratingID)
		return nil
	}

	// Update existing rating - validate ownership first
	existingRating, err := r.FindByID(ctx, userID, ratingEntity.GetID())
	if err != nil {
		return err
	}
	if existingRating == nil {
		return ownership.ErrResourceNotFound
	}

	// Update rating
	query := `
		UPDATE shared_deck_ratings
		SET rating = $1, comment = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5
	`

	now := time.Now()
	model.UpdatedAt = now

		var comment interface{}
		if model.Comment.Valid {
			comment = model.Comment.String
		}

		result, err := r.db.ExecContext(ctx, query,
			model.Rating,
			comment,
			model.UpdatedAt,
			model.ID,
			userID,
		)

	if err != nil {
		return fmt.Errorf("failed to update shared deck rating: %w", err)
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

// FindByID finds a shared deck rating by ID, filtering by userID to ensure ownership
func (r *SharedDeckRatingRepository) FindByID(ctx context.Context, userID int64, id int64) (*shareddeckrating.SharedDeckRating, error) {
	query := `
		SELECT id, user_id, shared_deck_id, rating, comment, created_at, updated_at
		FROM shared_deck_ratings
		WHERE id = $1 AND user_id = $2
	`

	var model models.SharedDeckRatingModel
	var comment sql.NullString
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.SharedDeckID,
		&model.Rating,
		&comment,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	model.Comment = comment

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find shared deck rating: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.SharedDeckRatingToDomain(&model)
}

// FindByUserID finds all shared deck ratings by a user
func (r *SharedDeckRatingRepository) FindByUserID(ctx context.Context, userID int64) ([]*shareddeckrating.SharedDeckRating, error) {
	query := `
		SELECT id, user_id, shared_deck_id, rating, comment, created_at, updated_at
		FROM shared_deck_ratings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shared deck ratings by user ID: %w", err)
	}
	defer rows.Close()

	var ratings []*shareddeckrating.SharedDeckRating
	for rows.Next() {
		var model models.SharedDeckRatingModel
		var comment sql.NullString

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.SharedDeckID,
			&model.Rating,
			&comment,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared deck rating: %w", err)
		}

		model.Comment = comment

		ratingEntity, err := mappers.SharedDeckRatingToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert shared deck rating to domain: %w", err)
		}
		ratings = append(ratings, ratingEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shared deck ratings: %w", err)
	}

	return ratings, nil
}

// Update updates an existing shared deck rating, validating ownership
func (r *SharedDeckRatingRepository) Update(ctx context.Context, userID int64, id int64, ratingEntity *shareddeckrating.SharedDeckRating) error {
	return r.Save(ctx, userID, ratingEntity)
}

// Delete deletes a shared deck rating, validating ownership
func (r *SharedDeckRatingRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingRating, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingRating == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (shared_deck_ratings doesn't have soft delete)
	query := `DELETE FROM shared_deck_ratings WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete shared deck rating: %w", err)
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

// Exists checks if a shared deck rating exists and belongs to the user
func (r *SharedDeckRatingRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM shared_deck_ratings
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check shared deck rating existence: %w", err)
	}

	return exists, nil
}

// FindBySharedDeckID finds all ratings for a shared deck
func (r *SharedDeckRatingRepository) FindBySharedDeckID(ctx context.Context, sharedDeckID int64) ([]*shareddeckrating.SharedDeckRating, error) {
	query := `
		SELECT id, user_id, shared_deck_id, rating, comment, created_at, updated_at
		FROM shared_deck_ratings
		WHERE shared_deck_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sharedDeckID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shared deck ratings by shared deck ID: %w", err)
	}
	defer rows.Close()

	var ratings []*shareddeckrating.SharedDeckRating
	for rows.Next() {
		var model models.SharedDeckRatingModel
		var comment sql.NullString

		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.SharedDeckID,
			&model.Rating,
			&comment,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared deck rating: %w", err)
		}

		model.Comment = comment

		ratingEntity, err := mappers.SharedDeckRatingToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert shared deck rating to domain: %w", err)
		}
		ratings = append(ratings, ratingEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shared deck ratings: %w", err)
	}

	return ratings, nil
}

// FindByUserIDAndSharedDeckID finds a rating by user and shared deck (one rating per user per deck)
func (r *SharedDeckRatingRepository) FindByUserIDAndSharedDeckID(ctx context.Context, userID int64, sharedDeckID int64) (*shareddeckrating.SharedDeckRating, error) {
	query := `
		SELECT id, user_id, shared_deck_id, rating, comment, created_at, updated_at
		FROM shared_deck_ratings
		WHERE user_id = $1 AND shared_deck_id = $2
	`

	var model models.SharedDeckRatingModel
	var comment sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID, sharedDeckID).Scan(
		&model.ID,
		&model.UserID,
		&model.SharedDeckID,
		&model.Rating,
		&comment,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	model.Comment = comment

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found, but not an error
		}
		return nil, fmt.Errorf("failed to find shared deck rating by user and shared deck: %w", err)
	}

	return mappers.SharedDeckRatingToDomain(&model)
}

// Ensure SharedDeckRatingRepository implements ISharedDeckRatingRepository
var _ secondary.ISharedDeckRatingRepository = (*SharedDeckRatingRepository)(nil)

