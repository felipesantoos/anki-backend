package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/services/search"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/infra/database/mappers"
	"github.com/felipesantos/anki-backend/infra/database/models"
	"github.com/felipesantos/anki-backend/pkg/ownership"
)

// CardRepository implements ICardRepository using PostgreSQL
type CardRepository struct {
	db *sql.DB
}

// NewCardRepository creates a new CardRepository instance
func NewCardRepository(db *sql.DB) secondary.ICardRepository {
	return &CardRepository{
		db: db,
	}
}

// Save saves or updates a card in the database
func (r *CardRepository) Save(ctx context.Context, userID int64, cardEntity *card.Card) error {
	model := mappers.CardToModel(cardEntity)

	// Validate deck ownership before saving
	deckOwnershipQuery := `SELECT user_id FROM decks WHERE id = $1 AND deleted_at IS NULL`
	var deckUserID int64
	err := r.db.QueryRowContext(ctx, deckOwnershipQuery, model.DeckID).Scan(&deckUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ownership.ErrResourceNotFound
		}
		return fmt.Errorf("failed to validate deck ownership: %w", err)
	}
	if err := ownership.EnsureOwnership(userID, deckUserID); err != nil {
		return ownership.ErrResourceNotFound
	}

	if cardEntity.GetID() == 0 {
		// Insert new card
		query := `
			INSERT INTO cards (
				note_id, card_type_id, deck_id, home_deck_id, due, interval, ease, lapses, reps,
				state, position, flag, suspended, buried, stability, difficulty, last_review_at,
				created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
			RETURNING id
		`

		now := time.Now()
		if model.CreatedAt.IsZero() {
			model.CreatedAt = now
		}
		if model.UpdatedAt.IsZero() {
			model.UpdatedAt = now
		}

		var homeDeckID interface{}
		if model.HomeDeckID.Valid {
			homeDeckID = model.HomeDeckID.Int64
		}

		var stability interface{}
		if model.Stability.Valid {
			stability = model.Stability.Float64
		}

		var difficulty interface{}
		if model.Difficulty.Valid {
			difficulty = model.Difficulty.Float64
		}

		var lastReviewAt interface{}
		if model.LastReviewAt.Valid {
			lastReviewAt = model.LastReviewAt.Time
		}

		var cardID int64
		err := r.db.QueryRowContext(ctx, query,
			model.NoteID,
			model.CardTypeID,
			model.DeckID,
			homeDeckID,
			model.Due,
			model.Interval,
			model.Ease,
			model.Lapses,
			model.Reps,
			model.State,
			model.Position,
			model.Flag,
			model.Suspended,
			model.Buried,
			stability,
			difficulty,
			lastReviewAt,
			model.CreatedAt,
			model.UpdatedAt,
		).Scan(&cardID)
		if err != nil {
			return fmt.Errorf("failed to create card: %w", err)
		}

		cardEntity.SetID(cardID)
		return nil
	}

	// Update existing card - validate ownership first
	existingCard, err := r.FindByID(ctx, userID, cardEntity.GetID())
	if err != nil {
		return err
	}
	if existingCard == nil {
		return ownership.ErrResourceNotFound
	}

	// Update card
	query := `
		UPDATE cards
		SET note_id = $1, card_type_id = $2, deck_id = $3, home_deck_id = $4, due = $5,
			interval = $6, ease = $7, lapses = $8, reps = $9, state = $10, position = $11,
			flag = $12, suspended = $13, buried = $14, stability = $15, difficulty = $16,
			last_review_at = $17, updated_at = $18
		WHERE id = $19 AND EXISTS (
			SELECT 1 FROM decks WHERE decks.id = cards.deck_id AND decks.user_id = $20 AND decks.deleted_at IS NULL
		)
	`

	now := time.Now()
	model.UpdatedAt = now

	var homeDeckID interface{}
	if model.HomeDeckID.Valid {
		homeDeckID = model.HomeDeckID.Int64
	}

	var stability interface{}
	if model.Stability.Valid {
		stability = model.Stability.Float64
	}

	var difficulty interface{}
	if model.Difficulty.Valid {
		difficulty = model.Difficulty.Float64
	}

	var lastReviewAt interface{}
	if model.LastReviewAt.Valid {
		lastReviewAt = model.LastReviewAt.Time
	}

	result, err := r.db.ExecContext(ctx, query,
		model.NoteID,
		model.CardTypeID,
		model.DeckID,
		homeDeckID,
		model.Due,
		model.Interval,
		model.Ease,
		model.Lapses,
		model.Reps,
		model.State,
		model.Position,
		model.Flag,
		model.Suspended,
		model.Buried,
		stability,
		difficulty,
		lastReviewAt,
		model.UpdatedAt,
		model.ID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
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

// FindByID finds a card by ID, filtering by userID via deck ownership to ensure ownership
func (r *CardRepository) FindByID(ctx context.Context, userID int64, id int64) (*card.Card, error) {
	query := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval,
			c.ease, c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried,
			c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
	`

	var model models.CardModel
	var homeDeckID sql.NullInt64
	var stability sql.NullFloat64
	var difficulty sql.NullFloat64
	var lastReviewAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&model.ID,
		&model.NoteID,
		&model.CardTypeID,
		&model.DeckID,
		&homeDeckID,
		&model.Due,
		&model.Interval,
		&model.Ease,
		&model.Lapses,
		&model.Reps,
		&model.State,
		&model.Position,
		&model.Flag,
		&model.Suspended,
		&model.Buried,
		&stability,
		&difficulty,
		&lastReviewAt,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ownership.ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to find card: %w", err)
	}

	model.HomeDeckID = homeDeckID
	model.Stability = stability
	model.Difficulty = difficulty
	model.LastReviewAt = lastReviewAt

	return mappers.CardToDomain(&model)
}

// FindByDeckID finds all cards in a deck, validating deck ownership
func (r *CardRepository) FindByDeckID(ctx context.Context, userID int64, deckID int64) ([]*card.Card, error) {
	query := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval,
			c.ease, c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried,
			c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.deck_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find cards by deck ID: %w", err)
	}
	defer rows.Close()

	var cards []*card.Card
	for rows.Next() {
		var model models.CardModel
		var homeDeckID sql.NullInt64
		var stability sql.NullFloat64
		var difficulty sql.NullFloat64
		var lastReviewAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.NoteID,
			&model.CardTypeID,
			&model.DeckID,
			&homeDeckID,
			&model.Due,
			&model.Interval,
			&model.Ease,
			&model.Lapses,
			&model.Reps,
			&model.State,
			&model.Position,
			&model.Flag,
			&model.Suspended,
			&model.Buried,
			&stability,
			&difficulty,
			&lastReviewAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		model.HomeDeckID = homeDeckID
		model.Stability = stability
		model.Difficulty = difficulty
		model.LastReviewAt = lastReviewAt

		cardEntity, err := mappers.CardToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card to domain: %w", err)
		}
		cards = append(cards, cardEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// Update updates an existing card, validating ownership via deck
func (r *CardRepository) Update(ctx context.Context, userID int64, id int64, cardEntity *card.Card) error {
	return r.Save(ctx, userID, cardEntity)
}

// Delete deletes a card, validating ownership via deck
func (r *CardRepository) Delete(ctx context.Context, userID int64, id int64) error {
	// Validate ownership first
	existingCard, err := r.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existingCard == nil {
		return ownership.ErrResourceNotFound
	}

	// Hard delete (cards don't have soft delete)
	query := `
		DELETE FROM cards
		WHERE id = $1 AND EXISTS (
			SELECT 1 FROM decks WHERE decks.id = cards.deck_id AND decks.user_id = $2 AND decks.deleted_at IS NULL
		)
	`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
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

// Exists checks if a card exists and belongs to a user's deck
func (r *CardRepository) Exists(ctx context.Context, userID int64, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM cards c
			INNER JOIN decks d ON c.deck_id = d.id
			WHERE c.id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check card existence: %w", err)
	}

	return exists, nil
}

// FindByNoteID finds all cards generated from a specific note, validating ownership
func (r *CardRepository) FindByNoteID(ctx context.Context, userID int64, noteID int64) ([]*card.Card, error) {
	query := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval,
			c.ease, c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried,
			c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.note_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find cards by note ID: %w", err)
	}
	defer rows.Close()

	var cards []*card.Card
	for rows.Next() {
		var model models.CardModel
		var homeDeckID sql.NullInt64
		var stability sql.NullFloat64
		var difficulty sql.NullFloat64
		var lastReviewAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.NoteID,
			&model.CardTypeID,
			&model.DeckID,
			&homeDeckID,
			&model.Due,
			&model.Interval,
			&model.Ease,
			&model.Lapses,
			&model.Reps,
			&model.State,
			&model.Position,
			&model.Flag,
			&model.Suspended,
			&model.Buried,
			&stability,
			&difficulty,
			&lastReviewAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		model.HomeDeckID = homeDeckID
		model.Stability = stability
		model.Difficulty = difficulty
		model.LastReviewAt = lastReviewAt

		cardEntity, err := mappers.CardToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card to domain: %w", err)
		}
		cards = append(cards, cardEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// FindByNoteIDs finds all cards generated from multiple notes, validating ownership
// Returns only cards that belong to the user's decks (filters out unauthorized cards)
func (r *CardRepository) FindByNoteIDs(ctx context.Context, userID int64, noteIDs []int64) ([]*card.Card, error) {
	if len(noteIDs) == 0 {
		return []*card.Card{}, nil
	}

	query := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval,
			c.ease, c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried,
			c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.note_id = ANY($1) AND d.user_id = $2 AND d.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(noteIDs), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find cards by note IDs: %w", err)
	}
	defer rows.Close()

	var cards []*card.Card
	for rows.Next() {
		var model models.CardModel
		var homeDeckID sql.NullInt64
		var stability sql.NullFloat64
		var difficulty sql.NullFloat64
		var lastReviewAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.NoteID,
			&model.CardTypeID,
			&model.DeckID,
			&homeDeckID,
			&model.Due,
			&model.Interval,
			&model.Ease,
			&model.Lapses,
			&model.Reps,
			&model.State,
			&model.Position,
			&model.Flag,
			&model.Suspended,
			&model.Buried,
			&stability,
			&difficulty,
			&lastReviewAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		model.HomeDeckID = homeDeckID
		model.Stability = stability
		model.Difficulty = difficulty
		model.LastReviewAt = lastReviewAt

		cardEntity, err := mappers.CardToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card to domain: %w", err)
		}
		cards = append(cards, cardEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// FindDueCards finds cards that are due for review in a deck
func (r *CardRepository) FindDueCards(ctx context.Context, userID int64, deckID int64, dueTimestamp int64) ([]*card.Card, error) {
	query := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval,
			c.ease, c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried,
			c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.deck_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
			AND c.suspended = FALSE AND c.buried = FALSE
			AND (c.state = 'new' OR (c.state IN ('review', 'relearn') AND c.due <= $3))
		ORDER BY c.due ASC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID, userID, dueTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to find due cards: %w", err)
	}
	defer rows.Close()

	var cards []*card.Card
	for rows.Next() {
		var model models.CardModel
		var homeDeckID sql.NullInt64
		var stability sql.NullFloat64
		var difficulty sql.NullFloat64
		var lastReviewAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.NoteID,
			&model.CardTypeID,
			&model.DeckID,
			&homeDeckID,
			&model.Due,
			&model.Interval,
			&model.Ease,
			&model.Lapses,
			&model.Reps,
			&model.State,
			&model.Position,
			&model.Flag,
			&model.Suspended,
			&model.Buried,
			&stability,
			&difficulty,
			&lastReviewAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		model.HomeDeckID = homeDeckID
		model.Stability = stability
		model.Difficulty = difficulty
		model.LastReviewAt = lastReviewAt

		cardEntity, err := mappers.CardToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card to domain: %w", err)
		}
		cards = append(cards, cardEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// FindByState finds all cards with a specific state in a deck
func (r *CardRepository) FindByState(ctx context.Context, userID int64, deckID int64, state valueobjects.CardState) ([]*card.Card, error) {
	query := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval,
			c.ease, c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried,
			c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.deck_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
			AND c.state = $3 AND c.suspended = FALSE AND c.buried = FALSE
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID, userID, state.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find cards by state: %w", err)
	}
	defer rows.Close()

	var cards []*card.Card
	for rows.Next() {
		var model models.CardModel
		var homeDeckID sql.NullInt64
		var stability sql.NullFloat64
		var difficulty sql.NullFloat64
		var lastReviewAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.NoteID,
			&model.CardTypeID,
			&model.DeckID,
			&homeDeckID,
			&model.Due,
			&model.Interval,
			&model.Ease,
			&model.Lapses,
			&model.Reps,
			&model.State,
			&model.Position,
			&model.Flag,
			&model.Suspended,
			&model.Buried,
			&stability,
			&difficulty,
			&lastReviewAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		model.HomeDeckID = homeDeckID
		model.Stability = stability
		model.Difficulty = difficulty
		model.LastReviewAt = lastReviewAt

		cardEntity, err := mappers.CardToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card to domain: %w", err)
		}
		cards = append(cards, cardEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// CountByDeckAndState counts cards with a specific state in a deck
func (r *CardRepository) CountByDeckAndState(ctx context.Context, userID int64, deckID int64, state valueobjects.CardState) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE c.deck_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
			AND c.state = $3 AND c.suspended = FALSE AND c.buried = FALSE
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, deckID, userID, state.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cards by state: %w", err)
	}

	return count, nil
}

// MoveCards moves all cards from a source deck (including sub-decks) to a target deck
func (r *CardRepository) MoveCards(ctx context.Context, userID int64, srcDeckID int64, targetDeckID int64) error {
	query := `
		UPDATE cards
		SET deck_id = $1, updated_at = $2
		WHERE deck_id IN (
			WITH RECURSIVE tree AS (
				SELECT id FROM decks WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
				UNION ALL
				SELECT d.id FROM decks d JOIN tree t ON d.parent_id = t.id WHERE d.deleted_at IS NULL
			) SELECT id FROM tree
		)
	`

	_, err := r.db.ExecContext(ctx, query, targetDeckID, time.Now(), srcDeckID, userID)
	if err != nil {
		return fmt.Errorf("failed to move cards: %w", err)
	}

	return nil
}

// DeleteByDeckRecursive deletes all cards from a deck and its sub-decks
func (r *CardRepository) DeleteByDeckRecursive(ctx context.Context, userID int64, deckID int64) error {
	query := `
		DELETE FROM cards
		WHERE deck_id IN (
			WITH RECURSIVE tree AS (
				SELECT id FROM decks WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
				UNION ALL
				SELECT d.id FROM decks d JOIN tree t ON d.parent_id = t.id WHERE d.deleted_at IS NULL
			) SELECT id FROM tree
		)
	`

	_, err := r.db.ExecContext(ctx, query, deckID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete cards recursively: %w", err)
	}

	return nil
}

// FindByAdvancedSearch finds cards matching advanced search criteria
// Used for is:new, is:due, is:review, prop: filters
func (r *CardRepository) FindByAdvancedSearch(ctx context.Context, userID int64, query *search.SearchQuery) ([]*card.Card, error) {
	if query == nil {
		return []*card.Card{}, nil
	}

	// Build dynamic SQL query
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base condition: ownership via deck JOIN
	conditions = append(conditions, "d.user_id = $1")
	args = append(args, userID)
	argIndex++

	conditions = append(conditions, "d.deleted_at IS NULL")

	// Card state filters
	for _, state := range query.States {
		switch state {
		case "new":
			conditions = append(conditions, "c.state = 'new'")
		case "review":
			conditions = append(conditions, "c.state = 'review'")
		case "learn":
			conditions = append(conditions, "c.state IN ('learn', 'relearn')")
		case "suspended":
			conditions = append(conditions, "c.suspended = TRUE")
		case "buried":
			conditions = append(conditions, "c.buried = TRUE")
		case "due":
			// Cards that are due: due <= now and not suspended/buried
			now := time.Now().Unix() * 1000 // Convert to milliseconds
			conditions = append(conditions, fmt.Sprintf("c.due <= $%d AND c.suspended = FALSE AND c.buried = FALSE", argIndex))
			args = append(args, now)
			argIndex++
		case "marked":
			// Join with notes table to check marked
			conditions = append(conditions, "EXISTS (SELECT 1 FROM notes n WHERE n.id = c.note_id AND n.marked = TRUE AND n.user_id = $1)")
		}
	}

	// Flag filters
	if len(query.Flags) > 0 {
		flagConditions := make([]string, len(query.Flags))
		for i, flag := range query.Flags {
			flagConditions[i] = fmt.Sprintf("c.flag = $%d", argIndex)
			args = append(args, flag)
			argIndex++
		}
		conditions = append(conditions, "("+strings.Join(flagConditions, " OR ")+")")
	}

	// Property filters
	for _, propFilter := range query.PropertyFilters {
		switch propFilter.Property {
		case "ivl":
			// Interval filter
			val, err := strconv.Atoi(propFilter.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid interval value: %s", propFilter.Value)
			}
			conditions = append(conditions, fmt.Sprintf("c.interval %s $%d", propFilter.Operator, argIndex))
			args = append(args, val)
			argIndex++
		case "due":
			// Due date filter (relative days)
			val, err := strconv.Atoi(propFilter.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid due value: %s", propFilter.Value)
			}
			// Calculate target timestamp (val days from now, in milliseconds)
			now := time.Now()
			targetTime := now.AddDate(0, 0, val)
			targetTimestamp := targetTime.Unix() * 1000
			conditions = append(conditions, fmt.Sprintf("c.due %s $%d", propFilter.Operator, argIndex))
			args = append(args, targetTimestamp)
			argIndex++
		case "lapses":
			// Lapses filter
			val, err := strconv.Atoi(propFilter.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid lapses value: %s", propFilter.Value)
			}
			conditions = append(conditions, fmt.Sprintf("c.lapses %s $%d", propFilter.Operator, argIndex))
			args = append(args, val)
			argIndex++
		case "reps":
			// Reps filter
			val, err := strconv.Atoi(propFilter.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid reps value: %s", propFilter.Value)
			}
			conditions = append(conditions, fmt.Sprintf("c.reps %s $%d", propFilter.Operator, argIndex))
			args = append(args, val)
			argIndex++
		}
	}

	// Build final query
	baseQuery := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval, c.ease, 
		       c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried, 
		       c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
	`

	whereClause := strings.Join(conditions, " AND ")
	queryStr := baseQuery + " WHERE " + whereClause + " ORDER BY c.created_at DESC"

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find cards by advanced search: %w", err)
	}
	defer rows.Close()

	return r.scanCards(rows)
}

// scanCards scans rows into card entities
func (r *CardRepository) scanCards(rows *sql.Rows) ([]*card.Card, error) {
	var cards []*card.Card
	for rows.Next() {
		var model models.CardModel
		var homeDeckID sql.NullInt64
		var stability sql.NullFloat64
		var difficulty sql.NullFloat64
		var lastReviewAt sql.NullTime

		err := rows.Scan(
			&model.ID,
			&model.NoteID,
			&model.CardTypeID,
			&model.DeckID,
			&homeDeckID,
			&model.Due,
			&model.Interval,
			&model.Ease,
			&model.Lapses,
			&model.Reps,
			&model.State,
			&model.Position,
			&model.Flag,
			&model.Suspended,
			&model.Buried,
			&stability,
			&difficulty,
			&lastReviewAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		model.HomeDeckID = homeDeckID
		model.Stability = stability
		model.Difficulty = difficulty
		model.LastReviewAt = lastReviewAt

		cardEntity, err := mappers.CardToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card to domain: %w", err)
		}
		cards = append(cards, cardEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return cards, nil
}

// FindAll finds cards for a user based on filters and pagination
func (r *CardRepository) FindAll(ctx context.Context, userID int64, filters card.CardFilters) ([]*card.Card, int, error) {
	// Build dynamic SQL query
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base condition: ownership via deck JOIN
	conditions = append(conditions, "d.user_id = $1")
	args = append(args, userID)
	argIndex++

	conditions = append(conditions, "d.deleted_at IS NULL")

	// Apply optional filters
	if filters.DeckID != nil {
		conditions = append(conditions, fmt.Sprintf("c.deck_id = $%d", argIndex))
		args = append(args, *filters.DeckID)
		argIndex++
	}

	if filters.State != nil {
		conditions = append(conditions, fmt.Sprintf("c.state = $%d", argIndex))
		args = append(args, *filters.State)
		argIndex++
	}

	if filters.Flag != nil {
		conditions = append(conditions, fmt.Sprintf("c.flag = $%d", argIndex))
		args = append(args, *filters.Flag)
		argIndex++
	}

	if filters.Suspended != nil {
		conditions = append(conditions, fmt.Sprintf("c.suspended = $%d", argIndex))
		args = append(args, *filters.Suspended)
		argIndex++
	}

	if filters.Buried != nil {
		conditions = append(conditions, fmt.Sprintf("c.buried = $%d", argIndex))
		args = append(args, *filters.Buried)
		argIndex++
	}

	// Build WHERE clause
	whereClause := strings.Join(conditions, " AND ")

	// Build COUNT query for total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE %s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count cards: %w", err)
	}

	// Build SELECT query with pagination
	baseQuery := `
		SELECT c.id, c.note_id, c.card_type_id, c.deck_id, c.home_deck_id, c.due, c.interval,
			c.ease, c.lapses, c.reps, c.state, c.position, c.flag, c.suspended, c.buried,
			c.stability, c.difficulty, c.last_review_at, c.created_at, c.updated_at
		FROM cards c
		INNER JOIN decks d ON c.deck_id = d.id
		WHERE %s
		ORDER BY c.created_at DESC
	`

	// Apply pagination
	if filters.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
		argIndex++
	}

	queryStr := fmt.Sprintf(baseQuery, whereClause)

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find cards: %w", err)
	}
	defer rows.Close()

	cards, err := r.scanCards(rows)
	if err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

// Ensure CardRepository implements ICardRepository
var _ secondary.ICardRepository = (*CardRepository)(nil)

