package export

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
)

// ExportService implements IExportService
type ExportService struct {
	deckRepo     secondary.IDeckRepository
	cardRepo     secondary.ICardRepository
	noteRepo     secondary.INoteRepository
	noteTypeRepo secondary.INoteTypeRepository
	mediaRepo    secondary.IMediaRepository
}

// NewExportService creates a new ExportService instance
func NewExportService(
	deckRepo secondary.IDeckRepository,
	cardRepo secondary.ICardRepository,
	noteRepo secondary.INoteRepository,
	noteTypeRepo secondary.INoteTypeRepository,
	mediaRepo secondary.IMediaRepository,
) primary.IExportService {
	return &ExportService{
		deckRepo:     deckRepo,
		cardRepo:     cardRepo,
		noteRepo:     noteRepo,
		noteTypeRepo: noteTypeRepo,
		mediaRepo:    mediaRepo,
	}
}

// ExportCollection exports all user data as a JSON snapshot in a stream
func (s *ExportService) ExportCollection(ctx context.Context, userID int64) (io.Reader, int64, error) {
	// 1. Fetch all decks
	decks, err := s.deckRepo.FindByUserID(ctx, userID, "")
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

// ExportNotes exports selected notes in the specified format
func (s *ExportService) ExportNotes(
	ctx context.Context,
	userID int64,
	noteIDs []int64,
	format string,
	includeMedia,
	includeScheduling bool,
) (io.Reader, int64, string, error) {
	if len(noteIDs) == 0 {
		return nil, 0, "", fmt.Errorf("note_ids cannot be empty")
	}

	// 1. Fetch notes with ownership validation
	notes, err := s.noteRepo.FindByIDs(ctx, userID, noteIDs)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to fetch notes: %w", err)
	}

	if len(notes) == 0 {
		return nil, 0, "", fmt.Errorf("no notes found or access denied")
	}

	// Validate that all requested notes were found (ownership check)
	foundIDs := make(map[int64]bool)
	for _, n := range notes {
		foundIDs[n.GetID()] = true
	}
	for _, id := range noteIDs {
		if !foundIDs[id] {
			return nil, 0, "", fmt.Errorf("note %d not found or access denied", id)
		}
	}

	// 2. Fetch cards if scheduling is included
	var cards []*card.Card
	if includeScheduling {
		cards, err = s.cardRepo.FindByNoteIDs(ctx, userID, noteIDs)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to fetch cards: %w", err)
		}
	}

	// 3. Fetch note types for the notes
	noteTypeIDs := make(map[int64]bool)
	for _, n := range notes {
		noteTypeIDs[n.GetNoteTypeID()] = true
	}

	noteTypes := make([]*notetype.NoteType, 0, len(noteTypeIDs))
	for noteTypeID := range noteTypeIDs {
		nt, err := s.noteTypeRepo.FindByID(ctx, userID, noteTypeID)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to fetch note type %d: %w", noteTypeID, err)
		}
		noteTypes = append(noteTypes, nt)
	}

	// 4. Fetch decks for the cards (if scheduling is included)
	var decks []*deck.Deck
	if includeScheduling {
		deckIDs := make(map[int64]bool)
		for _, c := range cards {
			deckIDs[c.GetDeckID()] = true
		}
		for deckID := range deckIDs {
			d, err := s.deckRepo.FindByID(ctx, userID, deckID)
			if err != nil {
				continue // Skip if deck not found
			}
			decks = append(decks, d)
		}
	}

	// 5. Extract and fetch media files if requested
	var mediaFiles []*media.Media
	if includeMedia {
		mediaFiles, err = s.extractMediaFromNotes(ctx, userID, notes)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to extract media: %w", err)
		}
	}

	// 6. Generate export based on format
	var reader io.Reader
	var size int64
	var filename string

	switch format {
	case "apkg":
		reader, size, err = GenerateAPKG(ctx, notes, cards, decks, noteTypes, mediaFiles, includeMedia)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to generate APKG: %w", err)
		}
		filename = "notes_export.apkg"
	case "text":
		reader, size, err = GenerateTextExport(notes, cards, includeScheduling)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to generate text export: %w", err)
		}
		filename = "notes_export.txt"
	default:
		return nil, 0, "", fmt.Errorf("unsupported format: %s", format)
	}

	return reader, size, filename, nil
}

// extractMediaFromNotes extracts media filenames from note fields and fetches media entities
func (s *ExportService) extractMediaFromNotes(ctx context.Context, userID int64, notes []*note.Note) ([]*media.Media, error) {
	mediaFilenames := make(map[string]bool)

	// Extract media references from note fields
	for _, n := range notes {
		var fields map[string]interface{}
		if err := json.Unmarshal([]byte(n.GetFieldsJSON()), &fields); err != nil {
			continue // Skip if fields can't be parsed
		}

		// Look for media references in field values (HTML img tags, audio tags, etc.)
		for _, val := range fields {
			valStr := fmt.Sprintf("%v", val)
			// Extract filenames from HTML tags (simplified regex)
			// Pattern: <img src="filename.jpg"> or <audio src="filename.mp3">
			// This is a simplified extraction - full implementation would use proper HTML parsing
			extracted := extractMediaFilenames(valStr)
			for _, filename := range extracted {
				mediaFilenames[filename] = true
			}
		}
	}

	// Fetch media entities by filename
	var mediaFiles []*media.Media
	for filename := range mediaFilenames {
		m, err := s.mediaRepo.FindByFilename(ctx, userID, filename)
		if err != nil || m == nil {
			continue // Skip if media not found
		}
		mediaFiles = append(mediaFiles, m)
	}

	return mediaFiles, nil
}

// extractMediaFilenames extracts media filenames from HTML content
// This is a simplified implementation - full version would use proper HTML parsing
func extractMediaFilenames(content string) []string {
	var filenames []string
	// Simple regex patterns for common media tags
	patterns := []string{
		`<img[^>]+src=["']([^"']+)["']`,
		`<audio[^>]+src=["']([^"']+)["']`,
		`<video[^>]+src=["']([^"']+)["']`,
		`\[sound:([^\]]+)\]`, // Anki sound syntax
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				filename := filepath.Base(match[1]) // Get just the filename
				if filename != "" {
					filenames = append(filenames, filename)
				}
			}
		}
	}

	return filenames
}

