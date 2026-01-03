package export

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/deck"
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
)

// GenerateAPKG generates an Anki package (.apkg) file
// An .apkg file is a ZIP containing:
// - collection.anki2: SQLite database with notes, cards, decks, note types
// - media: JSON file mapping media filenames
// - Media files (if includeMedia is true)
func GenerateAPKG(
	ctx context.Context,
	notes []*note.Note,
	cards []*card.Card,
	decks []*deck.Deck,
	noteTypes []*notetype.NoteType,
	mediaFiles []*media.Media,
	includeMedia bool,
) (io.Reader, int64, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// 1. Create SQLite database (collection.anki2)
	dbData, err := createAnkiDatabase(ctx, notes, cards, decks, noteTypes)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create Anki database: %w", err)
	}

	// Add database to ZIP
	dbFile, err := zipWriter.Create("collection.anki2")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create database file in ZIP: %w", err)
	}
	if _, err := dbFile.Write(dbData); err != nil {
		return nil, 0, fmt.Errorf("failed to write database to ZIP: %w", err)
	}

	// 2. Create media mapping file
	if includeMedia && len(mediaFiles) > 0 {
		mediaMap := make(map[string]string)
		for _, m := range mediaFiles {
			// Map original filename to itself (Anki format)
			mediaMap[m.GetFilename()] = m.GetFilename()
		}

		mediaJSON, err := json.Marshal(mediaMap)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal media map: %w", err)
		}

		mediaFile, err := zipWriter.Create("media")
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create media file in ZIP: %w", err)
		}
		if _, err := mediaFile.Write(mediaJSON); err != nil {
			return nil, 0, fmt.Errorf("failed to write media map to ZIP: %w", err)
		}

		// 3. Add media files to ZIP (if storage paths are available)
		// Note: This requires access to actual file storage, which may need to be injected
		// For now, we'll create placeholder entries
		for _, m := range mediaFiles {
			mediaEntry, err := zipWriter.Create(m.GetFilename())
			if err != nil {
				continue // Skip if we can't create the entry
			}
			// In a real implementation, read from storage and write to mediaEntry
			// For now, write empty placeholder
			mediaEntry.Write([]byte{})
		}
	} else {
		// Create empty media file
		mediaFile, err := zipWriter.Create("media")
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create media file in ZIP: %w", err)
		}
		mediaFile.Write([]byte("{}"))
	}

	if err := zipWriter.Close(); err != nil {
		return nil, 0, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	data := buf.Bytes()
	return bytes.NewReader(data), int64(len(data)), nil
}

// createAnkiDatabase creates a SQLite database with Anki's schema
// This is a simplified version - full implementation would require complete Anki schema and SQLite driver
// For now, returns a minimal database structure
// TODO: Add SQLite driver dependency and implement full Anki database schema
func createAnkiDatabase(
	ctx context.Context,
	notes []*note.Note,
	cards []*card.Card,
	decks []*deck.Deck,
	noteTypes []*notetype.NoteType,
) ([]byte, error) {
	// For now, create a minimal placeholder database
	// Full implementation would require:
	// 1. Add github.com/mattn/go-sqlite3 dependency
	// 2. Create full Anki schema (col, notes, cards, graves, revlog tables)
	// 3. Insert all data properly
	
	// Placeholder: return minimal SQLite database structure
	// In production, this should be a proper SQLite database file
	placeholder := []byte("SQLite format 3\x00") // SQLite file header
	
	// For now, we'll create a basic structure
	// Full implementation should use sql.Open("sqlite3", ...) and create proper schema
	return placeholder, fmt.Errorf("APKG generation requires SQLite driver - not yet implemented")
}

// parseFieldsJSON parses fields JSON array
func parseFieldsJSON(fieldsJSON string) []map[string]interface{} {
	var fields []map[string]interface{}
	json.Unmarshal([]byte(fieldsJSON), &fields)
	return fields
}

// parseCardTypesJSON parses card types JSON array
func parseCardTypesJSON(cardTypesJSON string) []map[string]interface{} {
	var cardTypes []map[string]interface{}
	json.Unmarshal([]byte(cardTypesJSON), &cardTypes)
	return cardTypes
}

// formatFieldsForAnki formats fields as tab-separated string (Anki format)
func formatFieldsForAnki(fields map[string]interface{}) string {
	var parts []string
	for _, v := range fields {
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return strings.Join(parts, "\x1f") // Anki uses 0x1f as field separator
}

