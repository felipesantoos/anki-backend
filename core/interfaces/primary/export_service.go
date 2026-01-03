package primary

import (
	"context"
	"io"
)

// IExportService defines the interface for exporting user data
type IExportService interface {
	// ExportCollection exports all user data (decks, cards, notes) as a stream
	ExportCollection(ctx context.Context, userID int64) (io.Reader, int64, error)

	// ExportNotes exports selected notes in the specified format
	// format: "apkg" for Anki package or "text" for plain text
	// includeMedia: whether to include media files referenced in notes
	// includeScheduling: whether to include card scheduling information
	// Returns: reader, size in bytes, filename, error
	ExportNotes(ctx context.Context, userID int64, noteIDs []int64, format string, includeMedia, includeScheduling bool) (io.Reader, int64, string, error)
}

