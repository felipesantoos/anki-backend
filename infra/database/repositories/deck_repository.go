package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
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
// The default deck is created with the name "Default" and standard configuration (empty JSON)
// Returns the deck ID or an error if creation fails
func (r *DeckRepository) CreateDefaultDeck(ctx context.Context, userID int64) (int64, error) {
	query := `
		INSERT INTO decks (user_id, name, parent_id, options_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	now := time.Now()
	defaultName := "Default"
	emptyOptions := "{}" // Empty JSON object for default options

	var deckID int64
	err := r.db.QueryRowContext(ctx, query,
		userID,
		defaultName,
		nil, // parent_id is NULL for root decks
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
func (r *DeckRepository) FindByID(ctx context.Context, userID int64, deckID int64) (*secondary.DeckData, error) {
	query := `
		SELECT id, user_id, name, parent_id, options_json, created_at, updated_at, deleted_at
		FROM decks
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var deck secondary.DeckData
	var createdAt, updatedAt, deletedAt sql.NullTime
	var parentID sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, deckID, userID).Scan(
		&deck.ID,
		&deck.UserID,
		&deck.Name,
		&parentID,
		&deck.OptionsJSON,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find deck: %w", err)
	}

	// Convert nullable fields
	if parentID.Valid {
		deck.ParentID = &parentID.Int64
	}
	deck.CreatedAt = createdAt
	deck.UpdatedAt = updatedAt
	deck.DeletedAt = deletedAt

	// Validate ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, deck.UserID); err != nil {
		return nil, ownership.ErrResourceNotFound
	}

	return &deck, nil
}

// FindByUserID finds all decks for a user
func (r *DeckRepository) FindByUserID(ctx context.Context, userID int64) ([]*secondary.DeckData, error) {
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

	var decks []*secondary.DeckData
	for rows.Next() {
		var deck secondary.DeckData
		var createdAt, updatedAt, deletedAt sql.NullTime
		var parentID sql.NullInt64

		err := rows.Scan(
			&deck.ID,
			&deck.UserID,
			&deck.Name,
			&parentID,
			&deck.OptionsJSON,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}

		if parentID.Valid {
			deck.ParentID = &parentID.Int64
		}
		deck.CreatedAt = createdAt
		deck.UpdatedAt = updatedAt
		deck.DeletedAt = deletedAt

		decks = append(decks, &deck)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decks: %w", err)
	}

	return decks, nil
}

// FindByParentID finds all decks with a specific parent ID, filtering by userID
func (r *DeckRepository) FindByParentID(ctx context.Context, userID int64, parentID int64) ([]*secondary.DeckData, error) {
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

	var decks []*secondary.DeckData
	for rows.Next() {
		var deck secondary.DeckData
		var createdAt, updatedAt, deletedAt sql.NullTime
		var parentIDVal sql.NullInt64

		err := rows.Scan(
			&deck.ID,
			&deck.UserID,
			&deck.Name,
			&parentIDVal,
			&deck.OptionsJSON,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}

		if parentIDVal.Valid {
			deck.ParentID = &parentIDVal.Int64
		}
		deck.CreatedAt = createdAt
		deck.UpdatedAt = updatedAt
		deck.DeletedAt = deletedAt

		decks = append(decks, &deck)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decks: %w", err)
	}

	return decks, nil
}

// Save creates or updates a deck
func (r *DeckRepository) Save(ctx context.Context, userID int64, deck *secondary.DeckData) error {
	if deck.ID == 0 {
		// Create new deck
		query := `
			INSERT INTO decks (user_id, name, parent_id, options_json, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		now := time.Now()
		var deckID int64
		err := r.db.QueryRowContext(ctx, query,
			userID,
			deck.Name,
			deck.ParentID,
			deck.OptionsJSON,
			now,
			now,
		).Scan(&deckID)

		if err != nil {
			return fmt.Errorf("failed to create deck: %w", err)
		}

		deck.ID = deckID
		deck.UserID = userID
		return nil
	}

	// Update existing deck - validate ownership first
	existingDeck, err := r.FindByID(ctx, userID, deck.ID)
	if err != nil {
		return err
	}

	// Ensure ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, existingDeck.UserID); err != nil {
		return ownership.ErrResourceNotFound
	}

	// Update deck
	query := `
		UPDATE decks
		SET name = $1, parent_id = $2, options_json = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		deck.Name,
		deck.ParentID,
		deck.OptionsJSON,
		now,
		deck.ID,
		userID,
	)

	if err != nil {
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
func (r *DeckRepository) Update(ctx context.Context, userID int64, deckID int64, deck *secondary.DeckData) error {
	return r.Save(ctx, userID, deck)
}

// Delete deletes a deck, validating ownership
func (r *DeckRepository) Delete(ctx context.Context, userID int64, deckID int64) error {
	// Validate ownership first
	existingDeck, err := r.FindByID(ctx, userID, deckID)
	if err != nil {
		return err
	}

	// Ensure ownership (defense in depth)
	if err := ownership.EnsureOwnership(userID, existingDeck.UserID); err != nil {
		return ownership.ErrResourceNotFound
	}

	// Soft delete
	query := `
		UPDATE decks
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, deckID, userID)
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

// Ensure DeckRepository implements IDeckRepository
var _ secondary.IDeckRepository = (*DeckRepository)(nil)
