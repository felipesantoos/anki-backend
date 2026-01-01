package primary

import (
	"context"

	"github.com/felipesantos/anki-backend/core/domain/entities/note_type"
)

// INoteTypeService defines the interface for note type management operations
type INoteTypeService interface {
	// Create creates a new note type
	Create(ctx context.Context, userID int64, name string, fieldsJSON string, cardTypesJSON string, templatesJSON string) (*notetype.NoteType, error)

	// FindByID finds a note type by ID
	FindByID(ctx context.Context, userID int64, id int64) (*notetype.NoteType, error)

	// FindByUserID finds all note types for a user, with optional search filter
	FindByUserID(ctx context.Context, userID int64, search string) ([]*notetype.NoteType, error)

	// Update updates an existing note type
	Update(ctx context.Context, userID int64, id int64, name string, fieldsJSON string, cardTypesJSON string, templatesJSON string) (*notetype.NoteType, error)

	// Delete deletes a note type (soft delete)
	Delete(ctx context.Context, userID int64, id int64) error
}

