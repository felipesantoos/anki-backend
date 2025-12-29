package secondary

import (
	"context"

	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
)

// INoteTypeRepository defines the interface for note type data persistence
// All methods that access specific resources require userID to ensure data isolation
type INoteTypeRepository interface {
	// Save saves or updates a note type in the database
	// If the note type has an ID, it updates the existing note type
	// If the note type has no ID, it creates a new note type and returns it with the ID set
	Save(ctx context.Context, userID int64, noteTypeEntity *notetype.NoteType) error

	// FindByID finds a note type by ID, filtering by userID to ensure ownership
	// Returns the note type if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*notetype.NoteType, error)

	// FindByUserID finds all note types for a user
	// Returns a list of note types belonging to the user
	FindByUserID(ctx context.Context, userID int64) ([]*notetype.NoteType, error)

	// Update updates an existing note type, validating ownership
	// Returns error if note type doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, noteTypeEntity *notetype.NoteType) error

	// Delete deletes a note type, validating ownership (soft delete)
	// Returns error if note type doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a note type exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByName finds a note type by name for a user
	FindByName(ctx context.Context, userID int64, name string) (*notetype.NoteType, error)

	// ExistsByName checks if a note type with the given name exists for the user
	ExistsByName(ctx context.Context, userID int64, name string) (bool, error)
}

