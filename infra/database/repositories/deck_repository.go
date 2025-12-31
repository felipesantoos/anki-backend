package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

var (
	// ErrDeckNotFound is returned when a deck is not found
	ErrDeckNotFound = errors.New("deck not found")
)

// DeckRepository implements IDeckRepository using PostgreSQL
type DeckRepository struct {
	db *sql.DB
}

// NewDeckRepository creates a new DeckRepository instance
func NewDeckRepository(db *sql.DB) secondary.IDeckRepository {
	return &DeckRepository{
		db: db,
	}
}

// CreateDefaultDeck creates a default deck for a user
func (r *DeckRepository) CreateDefaultDeck(ctx context.Context, userID int64) (int64, error) {
	query := `
		INSERT INTO decks (user_id, name, parent_id, options_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	now := time.Now()
	defaultName := "Default"
	emptyOptions := "{}"

	var deckID int64
	err := r.db.QueryRowContext(ctx, query,
		userID,
		defaultName,
		nil,
		emptyOptions,
		now,
		now,
	).Scan(&deckID)

	if err != nil {
		return 0, fmt.Errorf("failed to create default deck: %w", err)
	}

	return deckID, nil
}

// FindByID finds a deck by ID, filtering by userID to ensure ownership
func (r *DeckRepository) FindByID(ctx context.Context, userID int64, deckID int64) (*deck.Deck, error) {
	query := `
		SELECT id, user_id, name, parent_id, options_json, created_at, updated_at, deleted_at
		FROM decks
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var model models.DeckModel
	err := r.db.QueryRowContext(ctx, query, deckID, userID).Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&model.ParentID,
		&model.OptionsJSON,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find deck: %w", err)
	}

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, model.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return mappers.DeckToDomain(&model)
}

// FindByUserID finds all decks for a user
func (r *DeckRepository) FindByUserID(ctx context.Context, userID int64) ([]*deck.Deck, error) {
	query := `
		SELECT id, user_id, name, parent_id, options_json, created_at, updated_at, deleted_at
		FROM decks
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find decks by user ID: %w", err)
	}
	defer rows.Close()

	var decks []*deck.Deck
	for rows.Next() {
		var model models.DeckModel
		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.ParentID,
			&model.OptionsJSON,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}

		deckEntity, err := mappers.DeckToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert deck to domain: %w", err)
		}
		decks = append(decks, deckEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decks: %w", err)
	}

	return decks, nil
}

// FindByParentID finds all decks with a specific parent ID, filtering by userID
func (r *DeckRepository) FindByParentID(ctx context.Context, userID int64, parentID int64) ([]*deck.Deck, error) {
	query := `
		SELECT id, user_id, name, parent_id, options_json, created_at, updated_at, deleted_at
		FROM decks
		WHERE user_id = $1 AND parent_id = $2 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find decks by parent ID: %w", err)
	}
	defer rows.Close()

	var decks []*deck.Deck
	for rows.Next() {
		var model models.DeckModel
		err := rows.Scan(
			&model.ID,
			&model.UserID,
			&model.Name,
			&model.ParentID,
			&model.OptionsJSON,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}

		deckEntity, err := mappers.DeckToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert deck to domain: %w", err)
		}
		decks = append(decks, deckEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decks: %w", err)
	}

	return decks, nil
}

// Save creates or updates a deck
func (r *DeckRepository) Save(ctx context.Context, userID int64, deckEntity *deck.Deck) error {
	model := mappers.DeckToModel(deckEntity)

	if deckEntity.GetID() == 0 {
		// Create new deck
		query := `
			INSERT INTO decks (user_id, name, parent_id, options_json, created_at, updated_at)
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

		var deckID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			model.Name,
			model.ParentID,
			model.OptionsJSON,
			model.CreatedAt,
			model.UpdatedAt,
		).Scan(&deckID)

		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
				return fmt.Errorf("deck with name %s already exists at this level", model.Name)
			}
			return fmt.Errorf("failed to create deck: %w", err)
		}

		deckEntity.SetID(deckID)
		deckEntity.SetUserID(userID)
		return nil
	}

	// Update existing deck - validate ownership first
	existingDeck, err := r.FindByID(ctx, userID, deckEntity.GetID())
	if err != nil {
		return err
	}
	if existingDeck == nil {
		return ownership.ErrResourceNotFound
	}

	// Update deck
	query := `
		UPDATE decks
		SET name = $1, parent_id = $2, options_json = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	`

	now := time.Now()
	model.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		model.Name,
		model.ParentID,
		model.OptionsJSON,
		model.UpdatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return fmt.Errorf("deck with name %s already exists at this level", model.Name)
		}
		return fmt.Errorf("failed to update deck: %w", err)
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

// Update updates an existing deck, validating ownership
func (r *DeckRepository) Update(ctx context.Context, userID int64, deckID int64, deckEntity *deck.Deck) error {
	return r.Save(ctx, userID, deckEntity)
}

// Delete deletes a deck, validating ownership
func (r *DeckRepository) Delete(ctx context.Context, userID int64, deckID int64) error {
	// Validate ownership first
	existingDeck, err := r.FindByID(ctx, userID, deckID)
	if err != nil {
		return err
	}
	if existingDeck == nil {
		return ownership.ErrResourceNotFound
	}

	// Recursive Soft delete
	query := `
		WITH RECURSIVE deck_tree AS (
			SELECT id FROM decks WHERE id = $1 AND user_id = $2
			UNION ALL
			SELECT d.id FROM decks d JOIN deck_tree dt ON d.parent_id = dt.id
		)
		UPDATE decks
		SET deleted_at = $3
		WHERE id IN (SELECT id FROM deck_tree) AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, deckID, userID, now)
	if err != nil {
		return fmt.Errorf("failed to delete deck: %w", err)
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

// Exists checks if a deck with the given name exists for the user at the specified parent level
func (r *DeckRepository) Exists(ctx context.Context, userID int64, name string, parentID *int64) (bool, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT EXISTS(
				SELECT 1 FROM decks
				WHERE user_id = $1 AND name = $2 AND parent_id IS NULL AND deleted_at IS NULL
			)
		`
		args = []interface{}{userID, name}
	} else {
		query = `
			SELECT EXISTS(
				SELECT 1 FROM decks
				WHERE user_id = $1 AND name = $2 AND parent_id = $3 AND deleted_at IS NULL
			)
		`
		args = []interface{}{userID, name, *parentID}
	}

	var exists bool
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check deck existence: %w", err)
	}

	return exists, nil
}

// GetStats retrieves study statistics for a deck
func (r *DeckRepository) GetStats(ctx context.Context, userID int64, deckID int64) (*deck.DeckStats, error) {
	query := `
		SELECT
			ds.deck_id,
			ds.new_count,
			ds.learning_count,
			ds.review_count,
			ds.suspended_count,
			ds.notes_count,
			(
				SELECT COUNT(*)
				FROM cards c
				WHERE c.deck_id = ds.deck_id
				  AND c.due <= $1
				  AND c.suspended = FALSE
				  AND c.buried = FALSE
				  AND c.state IN ('learn', 'relearn', 'review')
			) as due_today_count
		FROM deck_statistics ds
		WHERE ds.deck_id = $2 AND ds.user_id = $3
	`

	now := time.Now().UnixMilli()
	var stats deck.DeckStats

	err := r.db.QueryRowContext(ctx, query, now, deckID, userID).Scan(
		&stats.DeckID,
		&stats.NewCount,
		&stats.LearningCount,
		&stats.ReviewCount,
		&stats.SuspendedCount,
		&stats.NotesCount,
		&stats.DueTodayCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to get deck statistics: %w", err)
	}

	return &stats, nil
}

// Ensure DeckRepository implements IDeckRepository
var _ secondary.IDeckRepository = (*DeckRepository)(nil)
