package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/review"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// ReviewRepository implements IReviewRepository using PostgreSQL
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository creates a new ReviewRepository instance
func NewReviewRepository(db *sql.DB) secondary.IReviewRepository {
	return &ReviewRepository{
		db: db,
	}
}

// Save saves or updates a review in the database
func (r *ReviewRepository) Save(ctx context.Context, userID int64, reviewEntity *review.Review) error {
	model := mappers.ReviewToModel(reviewEntity)

	// Validate card ownership via deck before saving
	cardOwnershipQuery := `
		SELECT d.user_id
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.id = $1 AND d.deleted_at IS NULL
	`
	var deckUserID int64
	err := r.db.QueryRowContext(ctx, cardOwnershipQuery, model.CardID).Scan(&deckUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ownership.ErrResourceNotFound
		}
		return fmt.Errorf("failed to validate card ownership: %w", err)
	}
	if err := ownership.EnsureOwnership(userID, deckUserID); err != nil {
		return ownership.ErrResourceNotFound
	}

	if reviewEntity.GetID() == 0 {
		// Insert new review
		query := `
			INSERT INTO reviews (card_id, rating, interval, ease, time_ms, type, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}

		var reviewID int64
		err := r.db.QueryRowContext(ctx, query,
			model.CardID,
			model.Rating,
			model.Interval,
			model.Ease,
			model.TimeMs,
			model.Type,
			model.CreatedAt,
		).Scan(&reviewID)
		if err != nil {
			return fmt.Errorf("failed to create review: %w", err)
		}

		reviewEntity.SetID(reviewID)
		return nil
	}

	// Update existing review - validate ownership first
	existingReview, err := r.FindByID(ctx, userID, reviewEntity.GetID())
	if err != nil {
		return err
	}
	if existingReview == nil {
		return ownership.ErrResourceNotFound
	}

	// Update review
	query := `
		UPDATE reviews
		SET card_id = $1, rating = $2, interval = $3, ease = $4, time_ms = $5, type = $6
		WHERE id = $7 AND EXISTS (
			SELECT 1 FROM cards c
			INNER JOIN decks d ON c.deck_id = d.id
			WHERE c.id = reviews.card_id AND d.user_id = $8 AND d.deleted_at IS NULL
		)
	`

	result, err := r.db.ExecContext(ctx, query,
		model.CardID,
		model.Rating,
		model.Interval,
		model.Ease,
		model.TimeMs,
		model.Type,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
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

// FindByID finds a review by ID, filtering by userID via card ownership to ensure ownership
func (r *ReviewRepository) FindByID(ctx context.Context, userID int64, id int64) (*review.Review, error) {
	query := `
		SELECT r.id, r.card_id, r.rating, r.interval, r.ease, r.time_ms, r.type, r.created_at
		FROM reviews r
		INNER JOIN cards c ON r.card_id = c.id
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE r.id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
	`

	var model models.ReviewModel
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.CardID,
		&model.Rating,
		&model.Interval,
		&model.Ease,
		&model.TimeMs,
		&model.Type,
		&model.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find review: %w", err)
	}

	return mappers.ReviewToDomain(&model)
}

// Update updates an existing review, validating ownership via card -> deck
func (r *ReviewRepository) Update(ctx context.Context, userID int64, id int64, reviewEntity *review.Review) error {
	return r.Save(ctx, userID, reviewEntity)
}

// Delete deletes a review, validating ownership via card -> deck
func (r *ReviewRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingReview, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingReview == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (reviews don't have soft delete)
	query := `
		DELETE FROM reviews
		WHERE id = $1 AND EXISTS (
			SELECT 1 FROM cards c
			INNER JOIN decks d ON c.deck_id = d.id
			WHERE c.id = reviews.card_id AND d.user_id = $2 AND d.deleted_at IS NULL
		)
	`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
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

// Exists checks if a review exists and belongs to a user's card
func (r *ReviewRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM reviews r
			INNER JOIN cards c ON r.card_id = c.id
			INNER JOIN decks d ON c.deck_id = d.id
			WHERE r.id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check review existence: %w", err)
	}

	return exists, nil
}

// FindByCardID finds all reviews for a specific card, validating ownership
func (r *ReviewRepository) FindByCardID(ctx context.Context, userID int64, cardID int64) ([]*review.Review, error) {
	query := `
		SELECT r.id, r.card_id, r.rating, r.interval, r.ease, r.time_ms, r.type, r.created_at
		FROM reviews r
		INNER JOIN cards c ON r.card_id = c.id
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE r.card_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, cardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by card ID: %w", err)
	}
	defer rows.Close()

	var reviews []*review.Review
	for rows.Next() {
		var model models.ReviewModel

		err := rows.Scan(
			&model.ID,
			&model.CardID,
			&model.Rating,
			&model.Interval,
			&model.Ease,
			&model.TimeMs,
			&model.Type,
			&model.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}

		reviewEntity, err := mappers.ReviewToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert review to domain: %w", err)
		}
		reviews = append(reviews, reviewEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviews: %w", err)
	}

	return reviews, nil
}

// FindByDateRange finds all reviews within a date range for a user
func (r *ReviewRepository) FindByDateRange(ctx context.Context, userID int64, startDate time.Time, endDate time.Time) ([]*review.Review, error) {
	query := `
		SELECT r.id, r.card_id, r.rating, r.interval, r.ease, r.time_ms, r.type, r.created_at
		FROM reviews r
		INNER JOIN cards c ON r.card_id = c.id
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE d.user_id = $1 AND d.deleted_at IS NULL
			AND r.created_at >= $2 AND r.created_at <= $3
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by date range: %w", err)
	}
	defer rows.Close()

	var reviews []*review.Review
	for rows.Next() {
		var model models.ReviewModel

		err := rows.Scan(
			&model.ID,
			&model.CardID,
			&model.Rating,
			&model.Interval,
			&model.Ease,
			&model.TimeMs,
			&model.Type,
			&model.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}

		reviewEntity, err := mappers.ReviewToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert review to domain: %w", err)
		}
		reviews = append(reviews, reviewEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviews: %w", err)
	}

	return reviews, nil
}

// DeleteByCardID deletes all reviews for a specific card, validating ownership
func (r *ReviewRepository) DeleteByCardID(ctx context.Context, userID int64, cardID int64) error {
	query := `
		DELETE FROM reviews
		WHERE card_id = $1 AND EXISTS (
			SELECT 1 FROM cards c
			INNER JOIN decks d ON c.deck_id = d.id
			WHERE c.id = reviews.card_id AND d.user_id = $2 AND d.deleted_at IS NULL
		)
	`

	_, err := r.db.ExecContext(ctx, query, cardID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete reviews by card ID: %w", err)
	}

	return nil
}

// Ensure ReviewRepository implements IReviewRepository
var _ secondary.IReviewRepository = (*ReviewRepository)(nil)

