package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
)

// GenerateTextExport generates a tab-separated text file from notes
// Format: GUID\tField1\tField2\t...\tTags
// If includeScheduling is true, includes card information as additional lines
func GenerateTextExport(notes []*note.Note, cards []*card.Card, includeScheduling bool) (io.Reader, int64, error) {
	var buf bytes.Buffer

	// Write header (optional, but helpful for import)
	buf.WriteString("GUID\t")

	// Determine field names from first note (if available)
	var fieldNames []string
	if len(notes) > 0 {
		var fields map[string]interface{}
		if err := json.Unmarshal([]byte(notes[0].GetFieldsJSON()), &fields); err == nil {
			for fieldName := range fields {
				fieldNames = append(fieldNames, fieldName)
			}
		}
	}

	// Write field names as header
	for _, fieldName := range fieldNames {
		buf.WriteString(fieldName)
		buf.WriteString("\t")
	}
	buf.WriteString("Tags\n")

	// Write notes
	for _, n := range notes {
		// GUID
		buf.WriteString(n.GetGUID().Value())
		buf.WriteString("\t")

		// Parse fields JSON
		var fields map[string]interface{}
		if err := json.Unmarshal([]byte(n.GetFieldsJSON()), &fields); err != nil {
			return nil, 0, fmt.Errorf("failed to parse fields JSON for note %d: %w", n.GetID(), err)
		}

		// Write field values in order
		for _, fieldName := range fieldNames {
			if val, ok := fields[fieldName]; ok {
				// Convert value to string, handling HTML content
				valStr := fmt.Sprintf("%v", val)
				// Escape tabs and newlines
				valStr = strings.ReplaceAll(valStr, "\t", " ")
				valStr = strings.ReplaceAll(valStr, "\n", " ")
				buf.WriteString(valStr)
			}
			buf.WriteString("\t")
		}

		// Tags
		tags := strings.Join(n.GetTags(), " ")
		buf.WriteString(tags)
		buf.WriteString("\n")

		// If includeScheduling, add card information
		if includeScheduling {
			noteCards := filterCardsByNoteID(cards, n.GetID())
			for _, c := range noteCards {
				buf.WriteString(fmt.Sprintf("CARD\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n",
					c.GetID(),
					c.GetDeckID(),
					c.GetDue(),
					c.GetInterval(),
					c.GetEase(),
					c.GetReps(),
					string(c.GetState()),
				))
			}
		}
	}

	data := buf.Bytes()
	return bytes.NewReader(data), int64(len(data)), nil
}

// filterCardsByNoteID filters cards by note ID
func filterCardsByNoteID(cards []*card.Card, noteID int64) []*card.Card {
	var result []*card.Card
	for _, c := range cards {
		if c.GetNoteID() == noteID {
			result = append(result, c)
		}
	}
	return result
}

