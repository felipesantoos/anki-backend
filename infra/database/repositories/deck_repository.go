package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
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

// Ensure DeckRepository implements IDeckRepository
var _ secondary.IDeckRepository = (*DeckRepository)(nil)
