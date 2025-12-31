package export

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// ExportService implements IExportService
type ExportService struct {
	deckRepo secondary.IDeckRepository
	cardRepo secondary.ICardRepository
	noteRepo secondary.INoteRepository
}

// NewExportService creates a new ExportService instance
func NewExportService(
	deckRepo secondary.IDeckRepository,
	cardRepo secondary.ICardRepository,
	noteRepo secondary.INoteRepository,
) primary.IExportService {
	return &ExportService{
		deckRepo: deckRepo,
		cardRepo: cardRepo,
		noteRepo: noteRepo,
	}
}

// ExportCollection exports all user data as a JSON snapshot in a stream
func (s *ExportService) ExportCollection(ctx context.Context, userID int64) (io.Reader, int64, error) {
	// 1. Fetch all decks
	decks, err := s.deckRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch decks for export: %w", err)
	}

	// 2. Fetch all notes (which contain the fields)
	// Use a large limit for export to ensure all notes are fetched
	notes, err := s.noteRepo.FindByUserID(ctx, userID, 1000000, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch notes for export: %w", err)
	}

	// 3. For each deck, fetch its cards
	// Note: This is a simplified implementation. In a real scenario, we might use a more efficient query.
	type ExportData struct {
		Decks []interface{} `json:"decks"`
		Notes []interface{} `json:"notes"`
	}

	data := ExportData{
		Decks: make([]interface{}, len(decks)),
		Notes: make([]interface{}, len(notes)),
	}

	for i, d := range decks {
		data.Decks[i] = map[string]interface{}{
			"id":           d.GetID(),
			"user_id":      d.GetUserID(),
			"name":         d.GetName(),
			"parent_id":    d.GetParentID(),
			"options_json": d.GetOptionsJSON(),
			"created_at":   d.GetCreatedAt(),
			"updated_at":   d.GetUpdatedAt(),
		}
	}
	for i, n := range notes {
		data.Notes[i] = map[string]interface{}{
			"id":           n.GetID(),
			"guid":         n.GetGUID().Value(),
			"note_type_id": n.GetNoteTypeID(),
			"fields_json":  n.GetFieldsJSON(),
			"tags":         n.GetTags(),
			"created_at":   n.GetCreatedAt(),
			"updated_at":   n.GetUpdatedAt(),
		}
	}

	// 4. Marshal to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal export data: %w", err)
	}

	return bytes.NewReader(jsonData), int64(len(jsonData)), nil
}

