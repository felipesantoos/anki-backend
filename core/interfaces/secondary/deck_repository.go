package secondary

import (
	"context"
)

// IDeckRepository defines the interface for deck data persistence
type IDeckRepository interface {
	// CreateDefaultDeck creates a default deck for a user
	// The default deck is created with the name "Default" and standard configuration
	// Returns the deck ID or an error if creation fails
	CreateDefaultDeck(ctx context.Context, userID int64) (int64, error)
}
