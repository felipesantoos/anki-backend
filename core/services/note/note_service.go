package note

import (
	"context"
	"fmt"
	"time"

	"github.com/felipesantos/anki-backend/core/domain/entities/card"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
	"github.com/felipesantos/anki-backend/core/interfaces/primary"
	"github.com/felipesantos/anki-backend/core/interfaces/secondary"
	"github.com/felipesantos/anki-backend/pkg/database"
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

	// Filter by DeckID (Highest priority)
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

