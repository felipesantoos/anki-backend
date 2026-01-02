package note

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	notetype "github.com/felipesantos/anki-backend/core/domain/entities/note_type"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/database"
	"github.com/felipesantos/anki-backend/pkg/ownership"
	"github.com/google/uuid"
)

// NoteService implements INoteService
type NoteService struct {
	noteRepo     secondary.INoteRepository
	cardRepo     secondary.ICardRepository
	noteTypeRepo secondary.INoteTypeRepository
	deckRepo     secondary.IDeckRepository
	tm           database.TransactionManager
}

// NewNoteService creates a new NoteService instance
func NewNoteService(
	noteRepo secondary.INoteRepository,
	cardRepo secondary.ICardRepository,
	noteTypeRepo secondary.INoteTypeRepository,
	deckRepo secondary.IDeckRepository,
	tm database.TransactionManager,
) primary.INoteService {
	return &NoteService{
		noteRepo:     noteRepo,
		cardRepo:     cardRepo,
		noteTypeRepo: noteTypeRepo,
		deckRepo:     deckRepo,
		tm:           tm,
	}
}

// Create creates a new note and generates associated cards
func (s *NoteService) Create(ctx context.Context, userID int64, noteTypeID int64, deckID int64, fieldsJSON string, tags []string) (*note.Note, error) {
	var noteEntity *note.Note

	err := s.tm.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Validate NoteType
		nt, err := s.noteTypeRepo.FindByID(txCtx, userID, noteTypeID)
		if err != nil {
			return err
		}
		if nt == nil {
			return fmt.Errorf("note type not found")
		}

		// 2. Validate Deck ownership
		d, err := s.deckRepo.FindByID(txCtx, userID, deckID)
		if err != nil {
			return err
		}
		if d == nil {
			return fmt.Errorf("deck not found")
		}

		// 3. Create Note
		guid, _ := valueobjects.NewGUID(uuid.New().String()) // Generate new GUID
		now := time.Now()
		noteEntity, err = note.NewBuilder().
			WithUserID(userID).
			WithGUID(guid).
			WithNoteTypeID(noteTypeID).
			WithFieldsJSON(fieldsJSON).
			WithTags(tags).
			WithCreatedAt(now).
			WithUpdatedAt(now).
			Build()
		if err != nil {
			return err
		}

		if err := s.noteRepo.Save(txCtx, userID, noteEntity); err != nil {
			return err
		}

		// 3. Generate Cards based on NoteType
		cardTypeCount := nt.GetCardTypeCount()
		for i := 0; i < cardTypeCount; i++ {
			cardEntity, err := card.NewBuilder().
				WithNoteID(noteEntity.GetID()).
				WithCardTypeID(i).
				WithDeckID(deckID).
				WithDue(now.Unix() * 1000). // New cards due now
				WithEase(2500).             // Default ease 250%
				WithState(valueobjects.CardStateNew).
				WithCreatedAt(now).
				WithUpdatedAt(now).
				Build()
			if err != nil {
				return err
			}

			if err := s.cardRepo.Save(txCtx, userID, cardEntity); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return noteEntity, nil
}

// FindByID finds a note by ID
func (s *NoteService) FindByID(ctx context.Context, userID int64, id int64) (*note.Note, error) {
	return s.noteRepo.FindByID(ctx, userID, id)
}

// FindAll finds notes for a user based on filters and pagination
func (s *NoteService) FindAll(ctx context.Context, userID int64, filters note.NoteFilters) ([]*note.Note, error) {
	// Set defaults for pagination
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Filter by Search (Highest priority)
	if filters.Search != "" {
		return s.noteRepo.FindBySearch(ctx, userID, filters.Search, filters.Limit, filters.Offset)
	}

	// Filter by DeckID
	if filters.DeckID != nil {
		return s.noteRepo.FindByDeckID(ctx, userID, *filters.DeckID, filters.Limit, filters.Offset)
	}

	// Filter by NoteTypeID
	if filters.NoteTypeID != nil {
		return s.noteRepo.FindByNoteTypeID(ctx, userID, *filters.NoteTypeID, filters.Limit, filters.Offset)
	}

	// Filter by Tags
	if len(filters.Tags) > 0 {
		return s.noteRepo.FindByTags(ctx, userID, filters.Tags, filters.Limit, filters.Offset)
	}

	// Default: Find all notes for user
	return s.noteRepo.FindByUserID(ctx, userID, filters.Limit, filters.Offset)
}

// Update updates an existing note
func (s *NoteService) Update(ctx context.Context, userID int64, id int64, fieldsJSON string, tags []string) (*note.Note, error) {
	existing, err := s.noteRepo.FindByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("note not found")
	}

	existing.SetFieldsJSON(fieldsJSON)
	existing.SetTags(tags)
	existing.SetUpdatedAt(time.Now())

	if err := s.noteRepo.Update(ctx, userID, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete deletes a note and its associated cards (soft delete)
func (s *NoteService) Delete(ctx context.Context, userID int64, id int64) error {
	return s.tm.WithTransaction(ctx, func(txCtx context.Context) error {
		// Soft delete note
		if err := s.noteRepo.Delete(txCtx, userID, id); err != nil {
			return err
		}

		// Soft delete cards
		cards, err := s.cardRepo.FindByNoteID(txCtx, userID, id)
		if err != nil {
			return err
		}

		for _, c := range cards {
			if err := s.cardRepo.Delete(txCtx, userID, c.GetID()); err != nil {
				return err
			}
		}

		return nil
	})
}

// AddTag adds a tag to a note
func (s *NoteService) AddTag(ctx context.Context, userID int64, id int64, tag string) error {
	existing, err := s.noteRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("note not found")
	}

	existing.AddTag(tag)
	return s.noteRepo.Update(ctx, userID, id, existing)
}

// RemoveTag removes a tag from a note
func (s *NoteService) RemoveTag(ctx context.Context, userID int64, id int64, tag string) error {
	existing, err := s.noteRepo.FindByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("note not found")
	}

	existing.RemoveTag(tag)
	return s.noteRepo.Update(ctx, userID, id, existing)
}

// Copy creates a copy of an existing note
func (s *NoteService) Copy(ctx context.Context, userID int64, noteID int64, deckID *int64, copyTags bool, copyMedia bool) (*note.Note, error) {
	var copiedNote *note.Note

	err := s.tm.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Find original note (validate ownership)
		originalNote, err := s.noteRepo.FindByID(txCtx, userID, noteID)
		if err != nil {
			// Convert ownership.ErrResourceNotFound to a more specific error message
			if errors.Is(err, ownership.ErrResourceNotFound) {
				return fmt.Errorf("note not found")
			}
			return err
		}
		if originalNote == nil {
			return fmt.Errorf("note not found")
		}

		// 2. Determine target deck
		var targetDeckID int64
		if deckID != nil {
			// Validate provided deck ownership
			d, err := s.deckRepo.FindByID(txCtx, userID, *deckID)
			if err != nil {
				// Convert ownership.ErrResourceNotFound to a more specific error message
				if errors.Is(err, ownership.ErrResourceNotFound) {
					return fmt.Errorf("deck not found")
				}
				return err
			}
			if d == nil {
				return fmt.Errorf("deck not found")
			}
			targetDeckID = *deckID
		} else {
			// Use original note's deck (get from first card)
			cards, err := s.cardRepo.FindByNoteID(txCtx, userID, noteID)
			if err != nil {
				return err
			}
			if len(cards) == 0 {
				return fmt.Errorf("note has no cards, cannot determine deck")
			}
			targetDeckID = cards[0].GetDeckID()
		}

		// 3. Get NoteType for card generation
		nt, err := s.noteTypeRepo.FindByID(txCtx, userID, originalNote.GetNoteTypeID())
		if err != nil {
			return err
		}
		if nt == nil {
			return fmt.Errorf("note type not found")
		}

		// 4. Prepare tags for copy
		tags := []string{} // Initialize as empty slice (PostgreSQL doesn't accept null for tags)
		if copyTags {
			tags = make([]string, len(originalNote.GetTags()))
			copy(tags, originalNote.GetTags())
		}

		// 5. Create new note entity
		guid, _ := valueobjects.NewGUID(uuid.New().String()) // Generate new GUID
		now := time.Now()
		copiedNote, err = note.NewBuilder().
			WithUserID(userID).
			WithGUID(guid).
			WithNoteTypeID(originalNote.GetNoteTypeID()).
			WithFieldsJSON(originalNote.GetFieldsJSON()).
			WithTags(tags).
			WithMarked(false). // Copy does not inherit marked status
			WithCreatedAt(now).
			WithUpdatedAt(now).
			Build()
		if err != nil {
			return err
		}

		// 6. Save new note
		if err := s.noteRepo.Save(txCtx, userID, copiedNote); err != nil {
			return err
		}

		// 7. Generate Cards based on NoteType (same logic as Create)
		cardTypeCount := nt.GetCardTypeCount()
		for i := 0; i < cardTypeCount; i++ {
			cardEntity, err := card.NewBuilder().
				WithNoteID(copiedNote.GetID()).
				WithCardTypeID(i).
				WithDeckID(targetDeckID).
				WithDue(now.Unix() * 1000). // New cards due now
				WithEase(2500).             // Default ease 250%
				WithState(valueobjects.CardStateNew).
				WithCreatedAt(now).
				WithUpdatedAt(now).
				Build()
			if err != nil {
				return err
			}

			if err := s.cardRepo.Save(txCtx, userID, cardEntity); err != nil {
				return err
			}
		}

		// 8. Handle media copying (placeholder for future implementation)
		// TODO: If copyMedia is true, parse fieldsJSON, extract media references,
		// copy media files via storage service, and update references in new note

		return nil
	})

	if err != nil {
		return nil, err
	}

	return copiedNote, nil
}

// FindDuplicates finds duplicate notes based on a field value
// If fieldName is empty and noteTypeID is provided, automatically uses the first field of the note type
func (s *NoteService) FindDuplicates(ctx context.Context, userID int64, noteTypeID *int64, fieldName string) (*note.DuplicateResult, error) {
	// If noteTypeID is provided, we can automatically detect the first field
	if noteTypeID != nil {
		nt, err := s.noteTypeRepo.FindByID(ctx, userID, *noteTypeID)
		if err != nil {
			if errors.Is(err, ownership.ErrResourceNotFound) {
				return nil, fmt.Errorf("note type not found")
			}
			return nil, err
		}
		if nt == nil {
			return nil, fmt.Errorf("note type not found")
		}

		// If fieldName is empty, automatically extract first field from note type
		if fieldName == "" {
			var err error
			fieldName, err = nt.GetFirstFieldName()
			if err != nil {
				return nil, err
			}
		} else {
			// Validate field name exists in note type
			if err := s.validateFieldName(nt, fieldName); err != nil {
				return nil, err
			}
		}
	} else {
		// If noteTypeID is not provided and fieldName is empty, return empty result
		if fieldName == "" {
			return &note.DuplicateResult{
				Duplicates: []*note.DuplicateGroup{},
				Total:      0,
			}, nil
		}
	}

	// Find duplicates via repository
	groups, err := s.noteRepo.FindDuplicatesByField(ctx, userID, noteTypeID, fieldName)
	if err != nil {
		return nil, err
	}

	return &note.DuplicateResult{
		Duplicates: groups,
		Total:      len(groups),
	}, nil
}

// FindDuplicatesByGUID finds duplicate notes based on GUID value
func (s *NoteService) FindDuplicatesByGUID(ctx context.Context, userID int64) (*note.DuplicateResult, error) {
	// Find duplicates via repository
	groups, err := s.noteRepo.FindDuplicatesByGUID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &note.DuplicateResult{
		Duplicates: groups,
		Total:      len(groups),
	}, nil
}

// validateFieldName validates that a field name exists in the note type
func (s *NoteService) validateFieldName(nt *notetype.NoteType, fieldName string) error {
	if nt.GetFieldsJSON() == "" {
		return fmt.Errorf("note type has no fields defined")
	}

	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(nt.GetFieldsJSON()), &fields); err != nil {
		return fmt.Errorf("invalid note type fields JSON: %w", err)
	}

	for _, field := range fields {
		// Check if field has "name" key
		if nameValue, exists := field["name"]; exists {
			if name, ok := nameValue.(string); ok && name == fieldName {
				return nil // Field found
			}
		}
	}

	return fmt.Errorf("field '%s' not found in note type", fieldName)
}

