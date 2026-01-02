package secondary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/services/search"
)

// INoteRepository defines the interface for note data persistence
// All methods that access specific resources require userID to ensure data isolation
type INoteRepository interface {
	// Save saves or updates a note in the database
	// If the note has an ID, it updates the existing note
	// If the note has no ID, it creates a new note and returns it with the ID set
	Save(ctx context.Context, userID int64, noteEntity *note.Note) error

	// FindByID finds a note by ID, filtering by userID to ensure ownership
	// Returns the note if found and belongs to user, nil if not found, or an error
	FindByID(ctx context.Context, userID int64, id int64) (*note.Note, error)

	// FindByUserID finds all notes for a user with pagination
	// Returns a list of notes belonging to the user
	FindByUserID(ctx context.Context, userID int64, limit int, offset int) ([]*note.Note, error)

	// FindByNoteTypeID finds all notes of a specific note type for a user with pagination
	FindByNoteTypeID(ctx context.Context, userID int64, noteTypeID int64, limit int, offset int) ([]*note.Note, error)

	// FindByDeckID finds all notes that have cards in a specific deck for a user with pagination
	FindByDeckID(ctx context.Context, userID int64, deckID int64, limit int, offset int) ([]*note.Note, error)

	// FindByTags finds all notes containing any of the specified tags for a user with pagination
	FindByTags(ctx context.Context, userID int64, tags []string, limit int, offset int) ([]*note.Note, error)

	// FindBySearch finds all notes containing the search text within fields_json for a user with pagination
	// Searches case-insensitively within JSON field values
	FindBySearch(ctx context.Context, userID int64, searchText string, limit int, offset int) ([]*note.Note, error)

	// Update updates an existing note, validating ownership
	// Returns error if note doesn't exist or doesn't belong to user
	Update(ctx context.Context, userID int64, id int64, noteEntity *note.Note) error

	// Delete deletes a note, validating ownership (soft delete)
	// Returns error if note doesn't exist or doesn't belong to user
	Delete(ctx context.Context, userID int64, id int64) error

	// Exists checks if a note exists and belongs to the user
	Exists(ctx context.Context, userID int64, id int64) (bool, error)

	// FindByGUID finds a note by GUID, filtering by userID to ensure ownership
	FindByGUID(ctx context.Context, userID int64, guid string) (*note.Note, error)

	// FindByAdvancedSearch finds notes matching advanced search criteria
	// Combines multiple filters: deck, tags, fields, card states, properties
	FindByAdvancedSearch(ctx context.Context, userID int64, query *search.SearchQuery, limit int, offset int) ([]*note.Note, error)

	// FindDuplicatesByField finds duplicate notes grouped by field value
	// Returns groups of notes that have the same value for the specified field
	FindDuplicatesByField(ctx context.Context, userID int64, noteTypeID *int64, fieldName string) ([]*note.DuplicateGroup, error)

	// FindDuplicatesByGUID finds duplicate notes grouped by GUID value
	// Returns groups of notes that have the same GUID (should not happen in normal operation, but useful for data integrity checks)
	FindDuplicatesByGUID(ctx context.Context, userID int64) ([]*note.DuplicateGroup, error)
}
