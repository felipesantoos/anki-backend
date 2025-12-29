package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/media"
)

// IMediaRepository defines the interface for media data persistence
// All methods that access specific resources require userID to ensure data isolation
type IMediaRepository interface {
	// Save saves or updates a media in the database
	// If the media has an ID, it updates the existing media
	// If the media has no ID, it creates a new media and returns it with the ID set
	Save(ctx context.Context, userID int64, mediaEntity *media.Media) error

	// FindByID finds a media by ID, filtering by userID to ensure ownership
	// Returns the media if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*media.Media, error)

	// FindByUserID finds all media for a user
	// Returns a list of media belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*media.Media, error)

	// Update updates an existing media, validating ownership
	// Returns error if media doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, mediaEntity *media.Media) error

	// Delete deletes a media, validating ownership (soft delete)
	// Returns error if media doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a media exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByHash finds a media by hash for a user
	FindByHash(ctx context.Context, userID int64, hash string) (*media.Media, error)

	// FindByFilename finds a media by filename for a user
	FindByFilename(ctx context.Context, userID int64, filename string) (*media.Media, error)
}

