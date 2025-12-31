package primary

import (
	"context"
	"io"
)

// IExportService defines the interface for exporting user data
type IExportService interface {
	// ExportCollection exports all user data (decks, cards, notes) as a stream
	ExportCollection(ctx context.Context, userID int64) (io.Reader, int64, error)
}

