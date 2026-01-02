package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/note"
)

// INoteService defines the interface for note management operations
type INoteService interface {
	// Create creates a new note and generates associated cards
	Create(ctx context.Context, userID int64, noteTypeID int64, deckID int64, fieldsJSON string, tags []string) (*note.Note, error)

	// FindByID finds a note by ID
	FindByID(ctx context.Context, userID int64, id int64) (*note.Note, error)

	// FindAll finds notes for a user based on filters and pagination
	FindAll(ctx context.Context, userID int64, filters note.NoteFilters) ([]*note.Note, error)

	// Update updates an existing note and its cards
	Update(ctx context.Context, userID int64, id int64, fieldsJSON string, tags []string) (*note.Note, error)

	// Delete deletes a note and its associated cards (soft delete)
	Delete(ctx context.Context, userID int64, id int64) error

	// AddTag adds a tag to a note
	AddTag(ctx context.Context, userID int64, id int64, tag string) error

	// RemoveTag removes a tag from a note
	RemoveTag(ctx context.Context, userID int64, id int64, tag string) error

	// Copy creates a copy of an existing note
	Copy(ctx context.Context, userID int64, noteID int64, deckID *int64, copyTags bool, copyMedia bool) (*note.Note, error)

	// FindDuplicates finds duplicate notes based on a field value
	FindDuplicates(ctx context.Context, userID int64, noteTypeID *int64, fieldName string) (*note.DuplicateResult, error)

	// FindDuplicatesByGUID finds duplicate notes based on GUID value
	// Returns groups of notes that have the same GUID (useful for data integrity checks)
	FindDuplicatesByGUID(ctx context.Context, userID int64) (*note.DuplicateResult, error)
}

